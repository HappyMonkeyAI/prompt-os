package execute

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	MountRoot   string
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
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer

	// Append command execution stdout/stderr logs to /tmp/promptos-install.log
	logFile, err := os.OpenFile("/tmp/promptos-install.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		defer logFile.Close()
		logFile.WriteString(fmt.Sprintf("\n=== Running: %s %s ===\n", name, strings.Join(args, " ")))
		cmd.Stdout = logFile
		cmd.Stderr = io.MultiWriter(&stderr, logFile)
	} else {
		cmd.Stderr = &stderr
	}

	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
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

	mountRoot := opts.MountRoot
	if mountRoot == "" {
		mountRoot = "/mnt/promptos-target"
	}
	efiPart := partitionPath(device, 1)
	rootPart := partitionPath(device, 2)
	if opts.SwapGiB > 0 {
		swapPart := partitionPath(device, 2)
		rootPart = partitionPath(device, 3)
		steps = append(steps,
			fmt.Sprintf("partprobe %s", device),
			fmt.Sprintf("mkfs.fat -F32 %s", efiPart),
			fmt.Sprintf("mkswap %s", swapPart),
			fmt.Sprintf("mkfs.ext4 -F %s", rootPart),
			fmt.Sprintf("mkdir -p %s", mountRoot),
			fmt.Sprintf("mount %s %s", rootPart, mountRoot),
			fmt.Sprintf("mkdir -p %s/boot/efi", mountRoot),
			fmt.Sprintf("mount %s %s/boot/efi", efiPart, mountRoot),
			fmt.Sprintf("swapon %s", swapPart),
		)
		return steps, nil
	}
	steps = append(steps,
		fmt.Sprintf("partprobe %s", device),
		fmt.Sprintf("mkfs.fat -F32 %s", efiPart),
		fmt.Sprintf("mkfs.ext4 -F %s", rootPart),
		fmt.Sprintf("mkdir -p %s", mountRoot),
		fmt.Sprintf("mount %s %s", rootPart, mountRoot),
		fmt.Sprintf("mkdir -p %s/boot/efi", mountRoot),
		fmt.Sprintf("mount %s %s/boot/efi", efiPart, mountRoot),
	)
	return steps, nil
}

func partitionPath(device string, number int) string {
	base := filepath.Base(device)
	if strings.HasPrefix(base, "nvme") || strings.HasPrefix(base, "mmcblk") {
		return fmt.Sprintf("%sp%d", device, number)
	}
	return fmt.Sprintf("%s%d", device, number)
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
