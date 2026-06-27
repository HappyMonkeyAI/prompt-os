package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ProviderDoneMsg is emitted when the user has confirmed all provider settings.
type ProviderDoneMsg struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// configPrompt describes a single text-input prompt within the provider config flow.
type configPrompt struct {
	label    string
	key      string // internal key: "api_key", "base_url", "model"
	dflt     string // shown as placeholder; used as value if user leaves empty
	password bool
	hint     string
}

// promptsForProvider returns the ordered list of config prompts for a provider.
func promptsForProvider(p string) []configPrompt {
	switch p {
	case "openai":
		return []configPrompt{
			{label: "API Key", key: "api_key", password: true,
				hint: "sk-…  (your OpenAI secret key)"},
		}
	case "openrouter":
		return []configPrompt{
			{label: "API Key", key: "api_key", password: true,
				hint: "sk-or-v1-…  (your OpenRouter key)"},
			{label: "Model", key: "model", dflt: "openai/gpt-4o-mini",
				hint: "e.g. openai/gpt-4o-mini · anthropic/claude-3-haiku · mistralai/mistral-7b"},
		}
	case "ollama":
		return []configPrompt{
			{label: "API Key (Optional)", key: "api_key", password: true,
				hint: "Optional key for authenticated local proxies/gateways (leave empty to skip)"},
			{label: "Base URL", key: "base_url", dflt: "http://localhost:11434",
				hint: "OpenAI-compatible endpoint — e.g. http://localhost:11434 (Ollama) · http://localhost:1234 (LM Studio)"},
			{label: "Model", key: "model", dflt: "llama3.2",
				hint: "e.g. llama3.2 · mistral · codellama · gemma2"},
		}
	}
	return nil
}

// ProviderModel handles AI provider selection and per-provider config entry.
type ProviderModel struct {
	providers    []string
	selected     int
	providerName string
	done         bool // true once all prompts answered; guards View() + Update()

	// dynamic prompts for the selected provider
	prompts     []configPrompt
	promptIndex int
	values      map[string]string // keyed by configPrompt.key
	input       textinput.Model
}

func NewProviderModel() ProviderModel {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 52

	return ProviderModel{
		providers: []string{"openai", "openrouter", "ollama"},
		selected:  0,
		values:    make(map[string]string),
		input:     ti,
	}
}

func (m ProviderModel) Init() tea.Cmd { return textinput.Blink }

func (m ProviderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// ── Phase 1: provider list ──────────────────────────────────────────
		if m.providerName == "" {
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
				m.providerName = m.providers[m.selected]
				m.prompts = promptsForProvider(m.providerName)
				m.promptIndex = 0
				m.values = make(map[string]string)
				m.setupInput(0)
				return m, textinput.Blink
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, cmd
		}

		// ── Phase 2: per-provider config prompts ────────────────────────────
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			val := m.input.Value()
			if val == "" {
				// Use the default if the user pressed Enter on an empty field
				val = m.prompts[m.promptIndex].dflt
			}
			m.values[m.prompts[m.promptIndex].key] = val
			m.promptIndex++

			if m.promptIndex >= len(m.prompts) {
				// Mark done before emitting — prevents View() panic while Bubble Tea drains events
				m.done = true
				return m, func() tea.Msg {
					return ProviderDoneMsg{
						Provider: m.providerName,
						APIKey:   m.values["api_key"],
						BaseURL:  m.values["base_url"],
						Model:    m.values["model"],
					}
				}
			}
			m.setupInput(m.promptIndex)
			return m, textinput.Blink
		}

		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m *ProviderModel) setupInput(idx int) {
	p := m.prompts[idx]
	m.input.SetValue("")
	m.input.Placeholder = p.hint
	m.input.CharLimit = 256
	m.input.Width = 52
	if p.password {
		m.input.EchoMode = textinput.EchoPassword
	} else {
		m.input.EchoMode = textinput.EchoNormal
	}
	m.input.Focus()
}

func (m ProviderModel) View() string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	muted  := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	box    := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 3).
		Width(60)

	// Guard: promptIndex may equal len(prompts) momentarily after the last Enter
	// before the parent AppModel processes ProviderDoneMsg and transitions stage.
	if m.done || (m.providerName != "" && m.promptIndex >= len(m.prompts)) {
		return box.Render(
			accent.Render("Provider configured!") + "\n\n" +
				muted.Render("Starting wizard…"),
		)
	}

	// Phase 1: provider list
	if m.providerName == "" {
		title := accent.Render("Select AI Provider") + "\n" +
			muted.Render("↑/↓ to navigate · Enter to select · q to quit") + "\n\n"

		descriptions := map[string]string{
			"openai":     "OpenAI  —  gpt-4o-mini (API key required)",
			"openrouter": "OpenRouter  —  200+ models via one key",
			"ollama":     "Ollama  —  local, no API key needed",
		}
		list := ""
		for i, p := range m.providers {
			label := descriptions[p]
			if i == m.selected {
				list += accent.Render("▶  "+label) + "\n"
			} else {
				list += "   " + label + "\n"
			}
		}
		return box.Render(title + list)
	}

	// Phase 2: config prompts
	prompt := m.prompts[m.promptIndex]
	progress := muted.Render(fmt.Sprintf("Provider: %s  ·  Field %d of %d",
		m.providerName, m.promptIndex+1, len(m.prompts)))
	header := accent.Render(prompt.label) + "\n"
	if prompt.dflt != "" {
		header += muted.Render("Default: "+prompt.dflt) + "\n"
	}
	header += "\n"
	hint := "Enter to confirm"
	if prompt.dflt != "" {
		hint = "Enter to confirm  (leave empty to use default)"
	}
	body := m.input.View() + "\n\n" + muted.Render(hint)

	return progress + "\n" + box.Render(header+body)
}

// WithProvider returns a copy of the model with the named provider set.
// Used in tests and for pre-seeding the model.
func (m ProviderModel) WithProvider(p string) ProviderModel { m.providerName = p; return m }

// ValidateKey returns true if the given key is non-empty and a provider is set.
// Pass an empty string to check the currently entered key.
func (m ProviderModel) ValidateKey(k string) bool {
	if k == "" {
		k = m.values["api_key"]
	}
	return len(k) > 0 && m.providerName != ""
}

// SetProvider is kept for backward compatibility (no-op; use WithProvider).
func (m ProviderModel) SetProvider(_ string) {}
