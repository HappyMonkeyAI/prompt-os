package triage

import (
	"context"
	"fmt"
)

type TriageAgent struct{}

func NewTriageAgent() *TriageAgent {
	return &TriageAgent{}
}

func (t *TriageAgent) Diagnose(ctx context.Context, symptom string) (string, error) {
	prompt := "You are an emergency Linux recovery assistant. Given the following boot or runtime failure, suggest safe recovery steps and commands.\n\nFailure: " + symptom
	return fmt.Sprintf("triage suggestion for: %s", prompt), nil
}
