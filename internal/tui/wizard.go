package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type WizardModel struct {
	step     int
	answers  map[string]string
	input    textinput.Model
	choices  []string
	selected int
}

func NewWizardModel() WizardModel {
	ti := textinput.New()
	ti.Focus()
	ti.Width = 50

	return WizardModel{
		step:    0,
		answers: make(map[string]string),
		input:   ti,
		choices: []string{},
	}
}

func (m WizardModel) Init() tea.Cmd { return textinput.Blink }

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(m.choices) > 0 {
				m.answers[m.currentKey()] = m.choices[m.selected]
			} else {
				m.answers[m.currentKey()] = m.input.Value()
			}
			m.step++
			m.resetInput()
			if m.step > 3 {
				return m, tea.Quit // done
			}
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.choices)-1 {
				m.selected++
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	if len(m.choices) == 0 {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m WizardModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Render(m.stepTitle())
	if len(m.choices) > 0 {
		list := ""
		for i, c := range m.choices {
			if i == m.selected {
				list += "> " + c + "\n"
			} else {
				list += "  " + c + "\n"
			}
		}
		return title + "\n\n" + list
	}
	return title + "\n\n" + m.input.View() + "\n(enter to continue, q to quit)"
}

func (m *WizardModel) resetInput() {
	m.input.SetValue("")
	m.choices = nil
	m.selected = 0
	switch m.step {
	case 0:
		m.choices = []string{"arch", "ubuntu", "debian"}
	case 1:
		m.choices = []string{"bleeding", "stable"}
	case 2:
		m.choices = []string{"wayland", "x11"}
	case 3:
		m.input.Placeholder = "e.g. nvidia / none"
	}
}

func (m WizardModel) stepTitle() string {
	switch m.step {
	case 0:
		return "Base distro?"
	case 1:
		return "Stability preference?"
	case 2:
		return "Display server?"
	case 3:
		return "GPU driver?"
	default:
		return "Wizard complete"
	}
}

func (m WizardModel) currentKey() string {
	switch m.step {
	case 0:
		return "base_distro"
	case 1:
		return "stability"
	case 2:
		return "display"
	case 3:
		return "gpu"
	default:
		return "done"
	}
}

func (m WizardModel) Answers() map[string]string { return m.answers }