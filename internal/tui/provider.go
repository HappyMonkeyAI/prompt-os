package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ProviderDoneMsg is emitted when the user has confirmed a provider + API key.
type ProviderDoneMsg struct {
	Provider string
	APIKey   string
}

// ProviderModel handles AI provider selection and API key entry.
type ProviderModel struct {
	provider  string
	keyInput  textinput.Model
	providers []string
	selected  int
	// enteringKey is true once a provider has been chosen and we're collecting the key
	enteringKey bool
}

func NewProviderModel() ProviderModel {
	ti := textinput.New()
	ti.Placeholder = "Paste your API key here"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 52
	ti.EchoMode = textinput.EchoPassword // mask the key

	return ProviderModel{
		providers: []string{"openai", "openrouter", "ollama"},
		selected:  0,
		keyInput:  ti,
	}
}

func (m ProviderModel) Init() tea.Cmd { return textinput.Blink }

func (m ProviderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.enteringKey {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				key := m.keyInput.Value()
				// Ollama needs no key; other providers require one
				if m.provider == "ollama" || len(key) > 0 {
					return m, func() tea.Msg {
						return ProviderDoneMsg{Provider: m.provider, APIKey: key}
					}
				}
				// Key required but empty — stay and flash placeholder
				m.keyInput.Placeholder = "Key required — paste it and press Enter"
				return m, nil
			}
			m.keyInput, cmd = m.keyInput.Update(msg)
			return m, cmd
		}

		// Provider selection mode
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
			m.enteringKey = true
			m.keyInput.SetValue("")
			m.keyInput.Focus()
			// Ollama needs no key — tell the user
			if m.provider == "ollama" {
				m.keyInput.Placeholder = "No key needed — press Enter to continue"
			} else {
				m.keyInput.Placeholder = "Paste your API key here"
			}
			return m, textinput.Blink
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m ProviderModel) View() string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	muted  := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	box    := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 3).
		Width(58)

	if m.enteringKey {
		hint := "Enter to confirm"
		if m.provider == "ollama" {
			hint = "Ensure Ollama is running locally, then press Enter"
		}
		content := accent.Render("Provider: "+m.provider) + "\n\n" +
			"API Key:\n" + m.keyInput.View() + "\n\n" +
			muted.Render(hint)
		return box.Render(content)
	}

	title := accent.Render("Select AI Provider") + "\n" +
		muted.Render("↑/↓ to navigate · Enter to select · q to quit") + "\n\n"

	list := ""
	descriptions := map[string]string{
		"openai":     "OpenAI  (gpt-4o-mini)",
		"openrouter": "OpenRouter  (access 200+ models)",
		"ollama":     "Ollama  (local, no API key needed)",
	}
	for i, p := range m.providers {
		label := descriptions[p]
		if label == "" {
			label = p
		}
		if i == m.selected {
			list += accent.Render("▶  "+label) + "\n"
		} else {
			list += "   " + label + "\n"
		}
	}

	return box.Render(title + list)
}

func (m ProviderModel) WithProvider(p string) ProviderModel { m.provider = p; return m }
func (m ProviderModel) SetProvider(p string)                 { /* no-op: use WithProvider for mutations */ }

func (m ProviderModel) ValidateKey(k string) bool {
	if k == "" {
		k = m.keyInput.Value()
	}
	return len(k) > 0 && m.provider != ""
}
