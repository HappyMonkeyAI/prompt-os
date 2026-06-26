package agent

import (
	"testing"
)

func TestBuildPlan(t *testing.T) {
	b := NewInstallerBuilder()
	plan, err := b.BuildPlan("test-installer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == "" {
		t.Fatal("expected non-empty plan")
	}
}
