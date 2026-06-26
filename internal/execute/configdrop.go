package execute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
)

var (
	ErrEmptyConfigs = errors.New("execute: blueprint has no configs to apply")
)

// ConfigDropOptions writes validated blueprint configs into a chroot mount.
type ConfigDropOptions struct {
	Blueprint *llm.Blueprint
	MountRoot string
	DryRun    bool
	Confirm   bool
}

// ConfigDropResult summarizes applied or planned file writes.
type ConfigDropResult struct {
	Mount   string
	DryRun  bool
	Files   []string
	Applied bool
}

type configWrite struct {
	AbsPath  string
	HostPath string
	Content  string
}

func buildConfigWrites(opts ConfigDropOptions) ([]configWrite, error) {
	if opts.Blueprint == nil {
		return nil, errors.New("execute: blueprint is required")
	}
	mount := strings.TrimSpace(opts.MountRoot)
	if mount == "" {
		return nil, ErrEmptyMountRoot
	}
	if len(opts.Blueprint.Configs) == 0 {
		return nil, ErrEmptyConfigs
	}

	mount = filepath.Clean(mount)
	absPaths := make([]string, 0, len(opts.Blueprint.Configs))
	for absPath := range opts.Blueprint.Configs {
		absPaths = append(absPaths, absPath)
	}
	sort.Strings(absPaths)

	writes := make([]configWrite, 0, len(absPaths))
	for _, absPath := range absPaths {
		if err := validateConfigTarget(absPath); err != nil {
			return nil, err
		}
		hostPath := filepath.Join(mount, strings.TrimPrefix(absPath, "/"))
		cleanHost, err := filepath.Abs(hostPath)
		if err != nil {
			return nil, err
		}
		if cleanHost != mount && !strings.HasPrefix(cleanHost, mount+string(os.PathSeparator)) {
			return nil, llm.ErrUnsafePath
		}
		writes = append(writes, configWrite{
			AbsPath:  absPath,
			HostPath: cleanHost,
			Content:  opts.Blueprint.Configs[absPath],
		})
	}
	return writes, nil
}

// BuildConfigDropPlan lists target file paths (host paths under MountRoot).
func BuildConfigDropPlan(opts ConfigDropOptions) ([]string, error) {
	writes, err := buildConfigWrites(opts)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(writes))
	for i, w := range writes {
		out[i] = w.HostPath
	}
	return out, nil
}

func validateConfigTarget(absPath string) error {
	if strings.Contains(absPath, "..") || !filepath.IsAbs(absPath) {
		return llm.ErrUnsafePath
	}
	if !strings.HasPrefix(absPath, "/etc/") && !strings.HasPrefix(absPath, "/opt/") {
		return llm.ErrUnsafePath
	}
	return nil
}

func fileModeForPath(absTarget string) os.FileMode {
	if strings.Contains(absTarget, "environment.d") || strings.Contains(absTarget, "ai-keys") {
		return 0o600
	}
	return 0o644
}

// ApplyConfigDrop writes blueprint config contents under MountRoot.
func ApplyConfigDrop(opts ConfigDropOptions) (ConfigDropResult, error) {
	writes, err := buildConfigWrites(opts)
	if err != nil {
		return ConfigDropResult{}, err
	}

	mount := filepath.Clean(strings.TrimSpace(opts.MountRoot))
	files := make([]string, len(writes))
	for i, w := range writes {
		files[i] = w.HostPath
	}

	result := ConfigDropResult{Mount: mount, DryRun: opts.DryRun, Files: files}
	if opts.DryRun {
		return result, nil
	}
	if !opts.Confirm {
		return ConfigDropResult{}, ErrConfirmRequired
	}

	for _, w := range writes {
		if err := os.MkdirAll(filepath.Dir(w.HostPath), 0o755); err != nil {
			return result, fmt.Errorf("execute: mkdir %s: %w", filepath.Dir(w.HostPath), err)
		}
		mode := fileModeForPath(w.AbsPath)
		if err := os.WriteFile(w.HostPath, []byte(w.Content), mode); err != nil {
			return result, fmt.Errorf("execute: write %s: %w", w.HostPath, err)
		}
	}

	result.Applied = true
	return result, nil
}