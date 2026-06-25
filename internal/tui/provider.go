package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type ProviderModel struct {
	provider   string
	keyInput   textinput.Model
	providers  []string
	selected   int
	validating bool
}

func NewProviderModel() ProviderModel {
	ti := textinput.New()
	ti.Placeholder = "sk-..."
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 40

	return ProviderModel{
		providers: []string{"openai", "anthropic", "gemini", "ollama"},
		selected:  0,
		keyInput:  ti,
	}
}

func (m *ProviderModel) Init() tea.Cmd { return textinput.Blink }

func (m *ProviderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.providers)-1 {
				m.selected++
			}
		case "enter":
			m.provider = m.providers[m.selected]
			m.validating = true
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	if m.validating {
		m.keyInput, cmd = m.keyInput.Update(msg)
	}
	return m, cmd
}

func (m ProviderModel) View() string {
	s := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	title := s.Render("Select Provider (↑/↓ + Enter)")

	list := ""
	for i, p := range m.providers {
		if i == m.selected {
			list += "> " + p + "\n"
		} else {
			list += "  " + p + "\n"
		}
	}

	if m.validating && m.provider != "" {
		return title + "\n\nProvider: " + m.provider + "\n\nAPI Key:\n" + m.keyInput.View() + "\n\n(Enter to confirm, q to quit)"
	}
	return title + "\n\n" + list + "\n(q to quit)"
}

func (m *ProviderModel) SetProvider(p string) { m.provider = p }
func (m ProviderModel) ValidateKey(k string) bool {
	if k == "" {
		k = m.keyInput.Value()
	}
	return len(k) > 0 && m.provider != ""
}