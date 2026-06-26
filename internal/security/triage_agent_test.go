package security

import (
	"context"
	"strings"
	"testing"
)

func TestDiagnoseReturnsSuggestion(t *testing.T) {
	agent := NewTriageAgent()
	out, err := agent.Diagnose(context.Background(), "kernel panic on boot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kernel panic on boot") {
		t.Fatalf("unexpected output: %s", out)
	}
}
