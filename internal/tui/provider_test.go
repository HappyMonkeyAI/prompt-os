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
	m2.values["api_key"] = "sk-test123"
	if !m2.ValidateKey("sk-test123") {
		t.Error("expected valid key to pass")
	}

	// Simulate flow: select provider with Enter → should set providerName
	m3 := NewProviderModel()
	updated, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 = updated.(ProviderModel)
	if m3.providerName == "" {
		t.Error("expected providerName to be set after Enter on provider list")
	}
	if len(m3.prompts) == 0 {
		t.Error("expected prompts to be populated after provider selection")
	}

	// Type a key value in the first prompt
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("sk-abc")}
	updated2, _ := m3.Update(msg)
	pm := updated2.(ProviderModel)
	if pm.input.Value() == "" {
		t.Error("expected input to capture runes during config prompt")
	}
}

func TestProviderConfigEnterWithOutOfRangePromptCompletes(t *testing.T) {
	m := NewProviderModel().WithProvider("openai")
	m.prompts = promptsForProvider("openai")
	m.values = make(map[string]string)
	m.promptIndex = len(m.prompts)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm := updated.(ProviderModel)
	if !pm.done {
		t.Fatal("expected model to be marked done")
	}
	if cmd == nil {
		t.Fatal("expected ProviderDoneMsg command")
	}

	msg := cmd()
	doneMsg, ok := msg.(ProviderDoneMsg)
	if !ok {
		t.Fatalf("expected ProviderDoneMsg, got %T", msg)
	}
	if doneMsg.Provider != "openai" {
		t.Fatalf("Provider = %q, want openai", doneMsg.Provider)
	}
}
