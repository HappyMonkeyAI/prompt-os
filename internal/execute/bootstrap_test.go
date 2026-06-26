package execute

import (
	"errors"
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