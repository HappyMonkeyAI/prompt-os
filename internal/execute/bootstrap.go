package execute

import (
	"errors"
	"fmt"
	"strings"

	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
)

var (
	ErrUnsupportedDistro = errors.New("execute: unsupported base_distro for bootstrap")
	ErrEmptyMountRoot    = errors.New("execute: mount root is required")
)

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

	for _, step := range steps {
		fields := strings.Fields(step)
		if len(fields) == 0 {
			continue
		}
		if err := runner.Run(fields[0], fields[1:]...); err != nil {
			return result, fmt.Errorf("execute: bootstrap failed on %q: %w", step, err)
		}
	}
	result.Applied = true
	return result, nil
}