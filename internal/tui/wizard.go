package tui

import (
	"strings"

	"github.com/HappyMonkeyAI/prompt-os/internal/hardware"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	recommended int      // choice index detected from hardware; -1 when none
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
		key:         "gpu",
		title:       "GPU / graphics driver",
		subtitle:    "Which GPU does this machine have?",
		choices:     []string{"nvidia", "amd", "intel", "none / virtual"},
		recommended: -1,
	},
}

// WizardModel walks the user through OS preference questions.
type WizardModel struct {
	step     int
	done     bool // true once all steps complete, guards View() and Update()
	answers  map[string]string
	input    textinput.Model
	selected int
	steps    []wizardStep
}

func NewWizardModel() WizardModel {
	return NewWizardModelWithHardware(hardware.HardwareInfo{})
}

func NewWizardModelWithHardware(hw hardware.HardwareInfo) WizardModel {
	ti := textinput.New()
	ti.Width = 60
	ti.CharLimit = 256

	m := WizardModel{
		step:    0,
		answers: make(map[string]string),
		input:   ti,
		steps:   stepsForHardware(hw),
	}
	m.resetForStep(0)
	return m
}

func (m WizardModel) Init() tea.Cmd { return textinput.Blink }

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Once done, ignore all further input
	if m.done {
		return m, nil
	}

	var cmd tea.Cmd
	step := m.steps[m.step]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			// Capture answer for current step
			if len(step.choices) > 0 {
				m.answers[step.key] = choiceValue(step.choices[m.selected])
			} else {
				val := strings.TrimSpace(m.input.Value())
				if val == "" {
					return m, nil // require non-empty free-text
				}
				m.answers[step.key] = val
			}

			m.step++
			if m.step >= len(m.steps) {
				// Mark done before emitting — prevents any further View/Update panics
				m.done = true
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

	// Only forward input to the text field when on a free-text step
	if !m.done && m.step < len(m.steps) && len(m.steps[m.step].choices) == 0 {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m WizardModel) View() string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	detected := lipgloss.NewStyle().Foreground(lipgloss.Color("#F6C177"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 3).
		Width(64)

	// Guard: step may momentarily equal len(wizardSteps) after the last Enter
	// before the parent AppModel processes WizardDoneMsg and transitions stage.
	if m.done || m.step >= len(m.steps) {
		return box.Render(
			accent.Render("Answers collected!") + "\n\n" +
				muted.Render("Connecting to AI provider…"),
		)
	}

	step := m.steps[m.step]
	progress := muted.Render("Step " + itoa(m.step+1) + " of " + itoa(len(m.steps)))

	header := accent.Render(step.title) + "\n" +
		muted.Render(step.subtitle) + "\n\n"

	var body string
	if len(step.choices) > 0 {
		for i, c := range step.choices {
			label := c
			if i == step.recommended {
				label = detected.Render(c)
			}
			if i == m.selected {
				body += accent.Render("▶  ") + label + "\n"
			} else {
				body += "   " + label + "\n"
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
	if step < len(m.steps) {
		m.input.Placeholder = m.steps[step].placeholder
		if m.steps[step].recommended >= 0 {
			m.selected = m.steps[step].recommended
		}
		if len(m.steps[step].choices) == 0 {
			m.input.Focus()
		} else {
			m.input.Blur()
		}
	}
}

func (m WizardModel) Answers() map[string]string { return m.answers }

func stepsForHardware(hw hardware.HardwareInfo) []wizardStep {
	steps := make([]wizardStep, len(wizardSteps))
	copy(steps, wizardSteps)
	for i := range steps {
		steps[i].choices = append([]string(nil), wizardSteps[i].choices...)
		if steps[i].recommended == 0 && steps[i].key != "use_case" {
			steps[i].recommended = -1
		}
		switch steps[i].key {
		case "gpu":
			steps[i].recommended = gpuChoiceIndex(hw.GPU)
			if steps[i].recommended >= 0 {
				steps[i].choices[steps[i].recommended] += "  (detected)"
			}
		}
	}

	if diskStep, ok := diskStepForHardware(hw); ok {
		insertAt := len(steps)
		for i, step := range steps {
			if step.key == "base_distro" {
				insertAt = i + 1
				break
			}
		}
		steps = append(steps[:insertAt], append([]wizardStep{diskStep}, steps[insertAt:]...)...)
	}
	return steps
}

func diskStepForHardware(hw hardware.HardwareInfo) (wizardStep, bool) {
	if len(hw.Disks) == 0 {
		return wizardStep{}, false
	}
	choices := make([]string, 0, len(hw.Disks)+1)
	recommended := -1
	for _, disk := range hw.Disks {
		if disk.Path == "" {
			continue
		}
		label := disk.Path
		if disk.Size != "" {
			label += "  " + disk.Size
		}
		if disk.Model != "" {
			label += " " + disk.Model
		}
		if disk.Transport != "" {
			label += " [" + disk.Transport + "]"
		}
		if disk.Removable {
			label += " removable"
		}
		if disk.ReadOnly {
			label += " read-only"
		}
		if recommended == -1 && !disk.Removable && !disk.ReadOnly {
			recommended = len(choices)
			label += "  (detected install target)"
		}
		choices = append(choices, label)
	}
	if len(choices) == 0 {
		return wizardStep{}, false
	}
	choices = append(choices, "skip disk prep")
	return wizardStep{
		key:         "target_disk",
		title:       "Install target disk",
		subtitle:    "Choose the disk PromptOS should wipe, partition, and mount",
		choices:     choices,
		recommended: recommended,
	}, true
}

func gpuChoiceIndex(gpu string) int {
	gpu = strings.ToLower(gpu)
	switch {
	case strings.Contains(gpu, "nvidia"):
		return 0
	case strings.Contains(gpu, "amd") || strings.Contains(gpu, "ati") || strings.Contains(gpu, "radeon"):
		return 1
	case strings.Contains(gpu, "intel"):
		return 2
	case strings.Contains(gpu, "vmware") || strings.Contains(gpu, "virtual") || strings.Contains(gpu, "virtio") || strings.Contains(gpu, "none"):
		return 3
	default:
		return -1
	}
}

func choiceValue(choice string) string {
	if idx := strings.Index(choice, "  "); idx >= 0 {
		return strings.TrimSpace(choice[:idx])
	}
	return strings.TrimSpace(choice)
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
