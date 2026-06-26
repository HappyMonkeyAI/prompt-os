package execute

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ErrInvalidDevice   = errors.New("execute: invalid block device path")
	ErrDeviceMounted   = errors.New("execute: refusing to prepare a mounted device")
	ErrConfirmRequired = errors.New("execute: real disk preparation requires ConfirmWipe")
	ErrEmptyDevice     = errors.New("execute: device is required")
	ErrRootDiskUnknown = errors.New("execute: cannot verify root disk")
)

// wholeDiskPattern matches common Linux whole-disk nodes (not partitions).
var wholeDiskPattern = regexp.MustCompile(`^/dev/(sd[a-z]+|vd[a-z]+|nvme[0-9]+n[0-9]+|mmcblk[0-9]+)$`)

// PrepareOptions controls target disk preparation.
type PrepareOptions struct {
	Device      string
	DryRun      bool
	ConfirmWipe bool
	EFIMiB      int // EFI partition size; 0 uses default 512
	SwapGiB     int // 0 disables swap partition
}

// PrepareResult summarizes planned or executed work.
type PrepareResult struct {
	Device  string
	DryRun  bool
	Steps   []string
	Applied bool
}

// CommandRunner executes shell commands (injectable for tests).
type CommandRunner interface {
	Output(name string, args ...string) ([]byte, error)
	Run(name string, args ...string) error
	Stat(path string) (os.FileInfo, error)
}

type osRunner struct{}

func (osRunner) Output(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func (osRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

func (osRunner) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// DefaultRunner uses the real OS.
var DefaultRunner CommandRunner = osRunner{}

// NormalizeDevice ensures a clean absolute device path.
func NormalizeDevice(device string) (string, error) {
	device = strings.TrimSpace(device)
	if device == "" {
		return "", ErrEmptyDevice
	}
	if !strings.HasPrefix(device, "/dev/") {
		device = filepath.Join("/dev", strings.TrimPrefix(device, "/dev/"))
	}
	device = filepath.Clean(device)
	if !wholeDiskPattern.MatchString(device) {
		return "", fmt.Errorf("%w: %s", ErrInvalidDevice, device)
	}
	return device, nil
}

// ValidateTargetDevice rejects unsafe targets such as the live root disk.
func ValidateTargetDevice(device string, runner CommandRunner) error {
	normalized, err := NormalizeDevice(device)
	if err != nil {
		return err
	}

	mounts, err := runner.Output("findmnt", "-n", "-o", "SOURCE", "--target", "/")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRootDiskUnknown, err)
	}
	rootSource := strings.TrimSpace(string(mounts))
	if rootSource == "" {
		return ErrRootDiskUnknown
	}
	if sameDisk(normalized, rootSource) {
		return fmt.Errorf("%w: %s is hosting /", ErrDeviceMounted, normalized)
	}
	rootDisk, err := parentDisk(rootSource, runner)
	if err != nil {
		return err
	}
	if rootDisk != "" && sameDisk(normalized, rootDisk) {
		return fmt.Errorf("%w: %s is hosting /", ErrDeviceMounted, normalized)
	}

	out, err := runner.Output("lsblk", "-ln", "-o", "NAME,MOUNTPOINT", normalized)
	if err != nil {
		return fmt.Errorf("execute: lsblk failed: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] != "" {
			return fmt.Errorf("%w: partition %s mounted at %s", ErrDeviceMounted, fields[0], fields[1])
		}
	}

	if _, err := runner.Stat(normalized); err != nil {
		return fmt.Errorf("execute: device not found: %w", err)
	}

	return nil
}

func parentDisk(device string, runner CommandRunner) (string, error) {
	out, err := runner.Output("lsblk", "-no", "PKNAME", device)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrRootDiskUnknown, err)
	}
	parent := strings.TrimSpace(string(out))
	if parent == "" {
		return "", nil
	}
	if strings.HasPrefix(parent, "/dev/") {
		return parent, nil
	}
	return filepath.Join("/dev", parent), nil
}

func sameDisk(whole, partOrWhole string) bool {
	whole = strings.TrimPrefix(whole, "/dev/")
	partOrWhole = strings.TrimPrefix(partOrWhole, "/dev/")
	if whole == partOrWhole {
		return true
	}
	return strings.HasPrefix(partOrWhole, whole)
}

// BuildPreparePlan returns the ordered steps for wiping + GPT layout.
func BuildPreparePlan(opts PrepareOptions) ([]string, error) {
	device, err := NormalizeDevice(opts.Device)
	if err != nil {
		return nil, err
	}
	efi := opts.EFIMiB
	if efi <= 0 {
		efi = 512
	}

	steps := []string{
		fmt.Sprintf("wipefs -a %s", device),
		fmt.Sprintf("parted -s %s mklabel gpt", device),
		fmt.Sprintf("parted -s %s mkpart ESP fat32 1MiB %dMiB", device, efi),
		fmt.Sprintf("parted -s %s set 1 esp on", device),
	}
	if opts.SwapGiB > 0 {
		steps = append(steps,
			fmt.Sprintf("parted -s %s mkpart primary linux-swap %dMiB %dGiB", device, efi, opts.SwapGiB),
			fmt.Sprintf("parted -s %s mkpart primary ext4 %dGiB 100%%", device, opts.SwapGiB),
		)
	} else {
		steps = append(steps,
			fmt.Sprintf("parted -s %s mkpart primary ext4 %dMiB 100%%", device, efi),
		)
	}
	return steps, nil
}

// PrepareDisk validates the target and either plans (dry-run) or executes steps.
func PrepareDisk(opts PrepareOptions, runner CommandRunner) (PrepareResult, error) {
	device, err := NormalizeDevice(opts.Device)
	if err != nil {
		return PrepareResult{}, err
	}
	if err := ValidateTargetDevice(device, runner); err != nil {
		return PrepareResult{}, err
	}

	steps, err := BuildPreparePlan(opts)
	if err != nil {
		return PrepareResult{}, err
	}

	result := PrepareResult{Device: device, DryRun: opts.DryRun, Steps: steps}
	if opts.DryRun {
		return result, nil
	}
	if !opts.ConfirmWipe {
		return PrepareResult{}, ErrConfirmRequired
	}

	for _, step := range steps {
		fields := strings.Fields(step)
		if len(fields) == 0 {
			continue
		}
		if err := runner.Run(fields[0], fields[1:]...); err != nil {
			return result, fmt.Errorf("execute: failed running %q: %w", step, err)
		}
	}
	result.Applied = true
	return result, nil
}
