package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// WizardDoneMsg is emitted when the user completes all wizard steps.
type WizardDoneMsg struct {
	Answers map[string]string
}

// wizardStep describes a single step in the wizard.
type wizardStep struct {
	key         string
	title       string
	subtitle    string
	choices     []string // nil → free-text input
	placeholder string   // used when choices == nil
}

var wizardSteps = []wizardStep{
	{
		key:         "use_case",
		title:       "What do you want this system for?",
		subtitle:    "Describe it in your own words",
		placeholder: "e.g. gaming PC, dev workstation, home media server, privacy-focused laptop…",
	},
	{
		key:      "base_distro",
		title:    "Base distribution",
		subtitle: "Choose the Linux base to build on",
		choices:  []string{"arch", "ubuntu", "debian"},
	},
	{
		key:      "stability",
		title:    "Stability preference",
		subtitle: "How cutting-edge do you want the packages?",
		choices:  []string{"stable  (recommended)", "bleeding  (latest everything)"},
	},
	{
		key:      "display",
		title:    "Display server",
		subtitle: "How should the desktop render graphics?",
		choices:  []string{"wayland  (modern, default)", "x11  (legacy, wider compat)", "none  (headless / server)"},
	},
	{
		key:      "gpu",
		title:    "GPU / graphics driver",
		subtitle: "Which GPU does this machine have?",
		choices:  []string{"nvidia", "amd", "intel", "none / virtual"},
	},
}

// WizardModel walks the user through OS preference questions.
type WizardModel struct {
	step    int
	answers map[string]string
	input   textinput.Model
	selected int
}

func NewWizardModel() WizardModel {
	ti := textinput.New()
	ti.Width = 60
	ti.CharLimit = 256

	m := WizardModel{
		step:    0,
		answers: make(map[string]string),
		input:   ti,
	}
	m.resetForStep(0)
	return m
}

func (m WizardModel) Init() tea.Cmd { return textinput.Blink }

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	step := wizardSteps[m.step]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			// Capture the answer for this step
			if len(step.choices) > 0 {
				// Strip the annotation after two spaces (e.g. "stable  (recommended)" → "stable")
				raw := step.choices[m.selected]
				if idx := strings.Index(raw, "  "); idx >= 0 {
					raw = raw[:idx]
				}
				m.answers[step.key] = raw
			} else {
				val := strings.TrimSpace(m.input.Value())
				if val == "" {
					return m, nil // require non-empty for free-text
				}
				m.answers[step.key] = val
			}

			m.step++
			if m.step >= len(wizardSteps) {
				answers := make(map[string]string, len(m.answers))
				for k, v := range m.answers {
					answers[k] = v
				}
				return m, func() tea.Msg { return WizardDoneMsg{Answers: answers} }
			}
			m.resetForStep(m.step)
			return m, textinput.Blink

		case "up", "k":
			if len(step.choices) > 0 && m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if len(step.choices) > 0 && m.selected < len(step.choices)-1 {
				m.selected++
			}
		}
	}

	if len(wizardSteps[m.step].choices) == 0 {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m WizardModel) View() string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	muted  := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	box    := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 3).
		Width(64)

	step := wizardSteps[m.step]
	progress := muted.Render("Step " + itoa(m.step+1) + " of " + itoa(len(wizardSteps)))

	header := accent.Render(step.title) + "\n" +
		muted.Render(step.subtitle) + "\n\n"

	var body string
	if len(step.choices) > 0 {
		for i, c := range step.choices {
			if i == m.selected {
				body += accent.Render("▶  "+c) + "\n"
			} else {
				body += "   " + c + "\n"
			}
		}
		body += "\n" + muted.Render("↑/↓ to choose · Enter to confirm")
	} else {
		body = m.input.View() + "\n\n" + muted.Render("Enter to confirm · q to quit")
	}

	return progress + "\n" + box.Render(header+body)
}

func (m *WizardModel) resetForStep(step int) {
	m.input.SetValue("")
	m.selected = 0
	if step < len(wizardSteps) {
		m.input.Placeholder = wizardSteps[step].placeholder
		if len(wizardSteps[step].choices) == 0 {
			m.input.Focus()
		} else {
			m.input.Blur()
		}
	}
}

func (m WizardModel) Answers() map[string]string { return m.answers }

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}