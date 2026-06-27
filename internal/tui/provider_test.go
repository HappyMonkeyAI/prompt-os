package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestProviderKeyValidation(t *testing.T) {
	m := NewProviderModel()

	// Empty key with no provider set should fail
	if m.ValidateKey("") {
		t.Error("expected empty key to fail validation")
	}

	// Valid key after choosing provider via WithProvider
	m2 := m.WithProvider("openai")
	if !m2.ValidateKey("sk-test123") {
		t.Error("expected valid key to pass")
	}

	// Simulate full flow: choose provider (enter), then type key
	m3 := NewProviderModel()
	// Press Enter to select the first provider → enters key-entry mode
	updated, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 = updated.(ProviderModel)
	if !m3.enteringKey {
		t.Error("expected enteringKey to be true after selecting provider")
	}
	// Type a key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("sk-abc")}
	updated2, _ := m3.Update(msg)
	pm := updated2.(ProviderModel)
	if pm.keyInput.Value() == "" {
		t.Error("expected key input to capture runes")
	}
}