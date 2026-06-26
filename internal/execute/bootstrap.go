package execute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
)

var (
	ErrUnsupportedDistro = errors.New("execute: unsupported base_distro for bootstrap")
	ErrEmptyMountRoot    = errors.New("execute: mount root is required")
	ErrUnsafePackageName = errors.New("execute: unsafe package name")
)

var packageNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9+._:@-]*$`)

// InstallOptions controls base system installation into a prepared chroot.
type InstallOptions struct {
	Blueprint  *llm.Blueprint
	MountRoot  string // e.g. /mnt/promptos-target
	DryRun     bool
	Confirm    bool
	ArchMirror string // optional override for pacstrap
}

// InstallResult captures planned or executed bootstrap work.
type InstallResult struct {
	Distro  string
	Mount   string
	DryRun  bool
	Steps   []string
	Applied bool
}

// BuildBootstrapPlan returns distro-specific commands to populate the target root.
func BuildBootstrapPlan(opts InstallOptions) ([]string, error) {
	if opts.Blueprint == nil {
		return nil, errors.New("execute: blueprint is required")
	}
	mount := strings.TrimSpace(opts.MountRoot)
	if mount == "" {
		return nil, ErrEmptyMountRoot
	}
	mount = strings.TrimSuffix(mount, "/")

	pkgs := append([]string(nil), opts.Blueprint.Packages...)
	if len(pkgs) == 0 {
		pkgs = defaultPackages(opts.Blueprint.BaseDistro)
	}
	if err := validatePackageNames(pkgs); err != nil {
		return nil, err
	}

	switch opts.Blueprint.BaseDistro {
	case "arch":
		base := []string{"base", "base-devel", "linux", "linux-firmware", "networkmanager", "sudo"}
		all := uniqueStrings(append(base, pkgs...))
		return []string{
			fmt.Sprintf("pacstrap -c %s %s", mount, strings.Join(all, " ")),
			fmt.Sprintf("genfstab -U %s >> /etc/fstab", mount),
			fmt.Sprintf("arch-chroot %s pacman -Syu --noconfirm", mount),
		}, nil
	case "debian", "ubuntu":
		suite := "bookworm"
		if opts.Blueprint.BaseDistro == "ubuntu" {
			suite = "noble"
		}
		if opts.Blueprint.StabilityPreference == "bleeding" && opts.Blueprint.BaseDistro == "debian" {
			suite = "sid"
		}
		steps := []string{
			fmt.Sprintf("debootstrap --arch=amd64 %s %s http://deb.debian.org/debian", suite, mount),
		}
		if len(pkgs) > 0 {
			steps = append(steps,
				fmt.Sprintf("chroot %s apt-get update", mount),
				fmt.Sprintf("chroot %s apt-get install -y %s", mount, strings.Join(pkgs, " ")),
			)
		}
		return steps, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDistro, opts.Blueprint.BaseDistro)
	}
}

func defaultPackages(distro string) []string {
	switch distro {
	case "arch":
		return []string{"vim", "git", "curl"}
	case "debian", "ubuntu":
		return []string{"vim", "git", "curl", "ca-certificates"}
	default:
		return nil
	}
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func validatePackageNames(pkgs []string) error {
	for _, pkg := range pkgs {
		pkg = strings.TrimSpace(pkg)
		if !packageNamePattern.MatchString(pkg) || strings.HasPrefix(pkg, "-") {
			return fmt.Errorf("%w: %q", ErrUnsafePackageName, pkg)
		}
	}
	return nil
}

// InstallBaseSystem plans or runs bootstrap commands for the blueprint distro.
func InstallBaseSystem(opts InstallOptions, runner CommandRunner) (InstallResult, error) {
	steps, err := BuildBootstrapPlan(opts)
	if err != nil {
		return InstallResult{}, err
	}

	result := InstallResult{
		Distro: opts.Blueprint.BaseDistro,
		Mount:  opts.MountRoot,
		DryRun: opts.DryRun,
		Steps:  steps,
	}
	if opts.DryRun {
		return result, nil
	}
	if !opts.Confirm {
		return InstallResult{}, ErrConfirmRequired
	}

	mount := strings.TrimSpace(opts.MountRoot)
	mount = strings.TrimSuffix(mount, "/")

	for _, step := range steps {
		if err := runStep(step, mount, runner); err != nil {
			return result, fmt.Errorf("execute: bootstrap failed on %q: %w", step, err)
		}
	}
	result.Applied = true
	return result, nil
}

func runStep(step string, mount string, runner CommandRunner) error {
	fields := strings.Fields(step)
	if len(fields) == 0 {
		return nil
	}

	// 1. Handle genfstab specially because of stdout redirection
	if fields[0] == "genfstab" {
		out, err := runner.Output("genfstab", "-U", mount)
		if err != nil {
			return fmt.Errorf("genfstab failed: %w", err)
		}
		fstabPath := filepath.Join(mount, "etc", "fstab")
		if err := os.MkdirAll(filepath.Dir(fstabPath), 0o755); err != nil {
			return fmt.Errorf("mkdir fstab: %w", err)
		}
		f, err := os.OpenFile(fstabPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("open fstab: %w", err)
		}
		defer f.Close()
		if _, err := f.Write(out); err != nil {
			return fmt.Errorf("write fstab: %w", err)
		}
		return nil
	}

	// 2. Handle pacstrap specially to ensure package names are passed correctly
	if fields[0] == "pacstrap" {
		if len(fields) < 4 || fields[1] != "-c" || fields[2] != mount {
			return fmt.Errorf("invalid pacstrap command: %q", step)
		}
		pkgs := fields[3:]
		args := append([]string{"-c", mount}, pkgs...)
		return runner.Run("pacstrap", args...)
	}

	// 3. Handle arch-chroot specially
	if fields[0] == "arch-chroot" {
		if len(fields) < 3 || fields[1] != mount {
			return fmt.Errorf("invalid arch-chroot command: %q", step)
		}
		cmdArgs := fields[2:]
		args := append([]string{mount}, cmdArgs...)
		return runner.Run("arch-chroot", args...)
	}

	// 4. Handle debootstrap specially
	if fields[0] == "debootstrap" {
		if len(fields) < 5 || fields[3] != mount {
			return fmt.Errorf("invalid debootstrap command: %q", step)
		}
		return runner.Run("debootstrap", fields[1:]...)
	}

	// 5. Handle chroot specially
	if fields[0] == "chroot" {
		if len(fields) < 3 || fields[1] != mount {
			return fmt.Errorf("invalid chroot command: %q", step)
		}
		cmdArgs := fields[2:]
		args := append([]string{mount}, cmdArgs...)
		return runner.Run("chroot", args...)
	}

	// Fallback
	return runner.Run(fields[0], fields[1:]...)
}
