package execute

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	rootSource string
	rootParent string
	lsblk      string
}

func (f fakeRunner) Output(name string, args ...string) ([]byte, error) {
	switch name {
	case "findmnt":
		return []byte(f.rootSource), nil
	case "lsblk":
		if len(args) >= 3 && args[0] == "-no" && args[1] == "PKNAME" {
			return []byte(f.rootParent), nil
		}
		return []byte(f.lsblk), nil
	default:
		return nil, errors.New("unexpected command")
	}
}

func (f fakeRunner) Run(name string, args ...string) error {
	return nil
}

func (f fakeRunner) Stat(path string) (fs.FileInfo, error) {
	return fakeFileInfo{name: path}, nil
}

type fakeFileInfo struct{ name string }

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() fs.FileMode  { return fs.ModeDevice }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) Sys() any           { return nil }

func TestNormalizeDeviceRejectsPartitions(t *testing.T) {
	_, err := NormalizeDevice("/dev/sda1")
	if !errors.Is(err, ErrInvalidDevice) {
		t.Fatalf("expected ErrInvalidDevice, got %v", err)
	}
}

func TestNormalizeDeviceAcceptsWholeDisk(t *testing.T) {
	got, err := NormalizeDevice("nvme0n1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/dev/nvme0n1" {
		t.Fatalf("expected /dev/nvme0n1, got %s", got)
	}
}

func TestValidateTargetDeviceRejectsRootDisk(t *testing.T) {
	runner := fakeRunner{rootSource: "/dev/sda2", lsblk: "sda1 \n"}
	err := ValidateTargetDevice("/dev/sda", runner)
	if !errors.Is(err, ErrDeviceMounted) {
		t.Fatalf("expected ErrDeviceMounted, got %v", err)
	}
}

func TestValidateTargetDeviceRejectsRootParentDisk(t *testing.T) {
	runner := fakeRunner{rootSource: "/dev/mapper/cryptroot", rootParent: "nvme0n1", lsblk: "nvme0n1 \n"}
	err := ValidateTargetDevice("/dev/nvme0n1", runner)
	if !errors.Is(err, ErrDeviceMounted) {
		t.Fatalf("expected ErrDeviceMounted, got %v", err)
	}
}

func TestBuildPreparePlanIncludesGPTAndEFI(t *testing.T) {
	steps, err := BuildPreparePlan(PrepareOptions{Device: "/dev/vdb"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(steps, "\n")
	for _, want := range []string{"wipefs -a /dev/vdb", "mklabel gpt", "mkpart ESP", "set 1 esp on"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing step fragment %q in:\n%s", want, joined)
		}
	}
}

func TestPrepareDiskDryRunDoesNotRequireConfirm(t *testing.T) {
	runner := fakeRunner{rootSource: "/dev/sda2", lsblk: "vdb1 \n"}
	res, err := PrepareDisk(PrepareOptions{Device: "/dev/vdb", DryRun: true}, runner)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if res.Applied {
		t.Fatal("dry-run should not apply changes")
	}
	if len(res.Steps) == 0 {
		t.Fatal("expected planned steps")
	}
}

func TestPrepareDiskRealRunRequiresConfirm(t *testing.T) {
	runner := fakeRunner{rootSource: "/dev/sda2", lsblk: "vdb1 \n"}
	_, err := PrepareDisk(PrepareOptions{Device: "/dev/vdb", DryRun: false}, runner)
	if !errors.Is(err, ErrConfirmRequired) {
		t.Fatalf("expected ErrConfirmRequired, got %v", err)
	}
}
