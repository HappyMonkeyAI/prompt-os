package execute

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
)

func TestBuildBootstrapPlanArch(t *testing.T) {
	bp := &llm.Blueprint{BaseDistro: "arch", Packages: []string{"htop"}}
	steps, err := BuildBootstrapPlan(InstallOptions{Blueprint: bp, MountRoot: "/mnt/target"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(steps, "\n")
	if !strings.Contains(joined, "pacstrap -c /mnt/target") || !strings.Contains(joined, "htop") {
		t.Fatalf("expected arch pacstrap plan, got:\n%s", joined)
	}
}

func TestBuildBootstrapPlanDebian(t *testing.T) {
	bp := &llm.Blueprint{BaseDistro: "debian", StabilityPreference: "stable"}
	steps, err := BuildBootstrapPlan(InstallOptions{Blueprint: bp, MountRoot: "/mnt/target"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(steps[0], "debootstrap") || !strings.Contains(steps[0], "bookworm") {
		t.Fatalf("unexpected debian plan: %v", steps)
	}
}

func TestBuildBootstrapPlanRejectsUnknownDistro(t *testing.T) {
	bp := &llm.Blueprint{BaseDistro: "gentoo"}
	_, err := BuildBootstrapPlan(InstallOptions{Blueprint: bp, MountRoot: "/mnt"})
	if !errors.Is(err, ErrUnsupportedDistro) {
		t.Fatalf("expected ErrUnsupportedDistro, got %v", err)
	}
}

func TestInstallBaseSystemDryRun(t *testing.T) {
	bp := &llm.Blueprint{BaseDistro: "ubuntu", StabilityPreference: "stable"}
	runner := fakeRunner{}
	res, err := InstallBaseSystem(InstallOptions{Blueprint: bp, MountRoot: "/mnt/target", DryRun: true}, runner)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if res.Applied || len(res.Steps) == 0 {
		t.Fatalf("expected dry-run steps only, got %+v", res)
	}
}

func TestInstallBaseSystemRealRunRequiresConfirm(t *testing.T) {
	bp := &llm.Blueprint{BaseDistro: "ubuntu"}
	runner := fakeRunner{}
	_, err := InstallBaseSystem(InstallOptions{Blueprint: bp, MountRoot: "/mnt/target"}, runner)
	if !errors.Is(err, ErrConfirmRequired) {
		t.Fatalf("expected ErrConfirmRequired, got %v", err)
	}
}

type mockRunner struct {
	runs    [][]string
	outputs map[string][]byte
}

func (m *mockRunner) Run(name string, args ...string) error {
	m.runs = append(m.runs, append([]string{name}, args...))
	return nil
}

func (m *mockRunner) Output(name string, args ...string) ([]byte, error) {
	m.runs = append(m.runs, append([]string{name}, args...))
	if out, ok := m.outputs[name]; ok {
		return out, nil
	}
	return nil, nil
}

func (m *mockRunner) Stat(path string) (os.FileInfo, error) {
	return nil, nil
}

func TestInstallBaseSystemArch(t *testing.T) {
	bp := &llm.Blueprint{BaseDistro: "arch", Packages: []string{"htop"}}
	mount := t.TempDir()

	runner := &mockRunner{
		outputs: map[string][]byte{
			"genfstab": []byte("# Test Fstab Output\n"),
		},
	}

	opts := InstallOptions{
		Blueprint: bp,
		MountRoot: mount,
		Confirm:   true,
	}

	res, err := InstallBaseSystem(opts, runner)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if !res.Applied {
		t.Fatalf("expected Applied to be true")
	}

	fstabPath := filepath.Join(mount, "etc", "fstab")
	fstabContent, err := os.ReadFile(fstabPath)
	if err != nil {
		t.Fatalf("failed to read fstab: %v", err)
	}
	if !strings.Contains(string(fstabContent), "# Test Fstab Output") {
		t.Fatalf("fstab content mismatch: %q", string(fstabContent))
	}

	expectedRuns := [][]string{
		{"pacstrap", "-c", mount, "base", "base-devel", "linux", "linux-firmware", "networkmanager", "sudo", "htop"},
		{"genfstab", "-U", mount},
		{"arch-chroot", mount, "pacman", "-Syu", "--noconfirm"},
	}

	if len(runner.runs) != len(expectedRuns) {
		t.Fatalf("expected %d runs, got %d", len(expectedRuns), len(runner.runs))
	}
	for i, run := range runner.runs {
		if strings.Join(run, " ") != strings.Join(expectedRuns[i], " ") {
			t.Errorf("run %d mismatch: got %v, want %v", i, run, expectedRuns[i])
		}
	}
}

func TestInstallBaseSystemDebian(t *testing.T) {
	bp := &llm.Blueprint{BaseDistro: "debian", StabilityPreference: "stable", Packages: []string{"htop"}}
	mount := t.TempDir()

	runner := &mockRunner{}
	opts := InstallOptions{
		Blueprint: bp,
		MountRoot: mount,
		Confirm:   true,
	}

	res, err := InstallBaseSystem(opts, runner)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if !res.Applied {
		t.Fatalf("expected Applied to be true")
	}

	expectedRuns := [][]string{
		{"debootstrap", "--arch=amd64", "bookworm", mount, "http://deb.debian.org/debian"},
		{"chroot", mount, "apt-get", "update"},
		{"chroot", mount, "apt-get", "install", "-y", "htop"},
	}

	if len(runner.runs) != len(expectedRuns) {
		t.Fatalf("expected %d runs, got %d", len(expectedRuns), len(runner.runs))
	}
	for i, run := range runner.runs {
		if strings.Join(run, " ") != strings.Join(expectedRuns[i], " ") {
			t.Errorf("run %d mismatch: got %v, want %v", i, run, expectedRuns[i])
		}
	}
}