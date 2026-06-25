package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestProviderKeyValidation(t *testing.T) {
	m := NewProviderModel()

	// Empty key should be invalid
	if m.ValidateKey("") {
		t.Error("expected empty key to fail validation")
	}

	// Valid key after choosing provider
	m.SetProvider("openai")
	if !m.ValidateKey("sk-test123") {
		t.Error("expected valid key to pass")
	}

	// Simulate full flow: choose + enter + key input
	m2 := NewProviderModel()
	// select first, enter
	_, _ = m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2.SetProvider("openai")
	m2.validating = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("sk-abc")}
	updated, _ := m2.Update(msg)
	if pm, ok := updated.(*ProviderModel); ok {
		if pm.keyInput.Value() == "" {
			t.Error("expected key input to capture runes")
		}
	}
}