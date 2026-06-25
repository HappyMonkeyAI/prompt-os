package tui

import (
	"context"
	"fmt"

	"github.com/HappyMonkeyAI/prompt-os/internal/hardware"
	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BlueprintModel handles the final step: calling the LLM and showing the generated blueprint.
type BlueprintModel struct {
	client    llm.LLMClient
	wizard    WizardModel
	hw        hardware.HardwareInfo
	blueprint *llm.Blueprint
	err       error
	loading   bool
}

func NewBlueprintModel(client llm.LLMClient, wizard WizardModel, hw hardware.HardwareInfo) BlueprintModel {
	return BlueprintModel{
		client:  client,
		wizard:  wizard,
		hw:      hw,
		loading: true,
	}
}

func (m BlueprintModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Build a simple prompt from wizard + hardware
		prompt := fmt.Sprintf(
			"User preferences: %v\nHardware: %+v\nGenerate a Linux installation blueprint.",
			m.wizard.Answers(), m.hw,
		)

		resp, err := m.client.Generate(context.Background(), prompt)
		if err != nil {
			return errMsg{err}
		}

		bp, err := llm.ValidateBlueprint([]byte(resp))
		if err != nil {
			return errMsg{err}
		}
		return blueprintMsg{bp}
	}
}

func (m BlueprintModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case blueprintMsg:
		m.blueprint = msg.bp
		m.loading = false
	case errMsg:
		m.err = msg.err
		m.loading = false
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m BlueprintModel) View() string {
	if m.loading {
		return "Generating blueprint with " + m.client.Name() + "...\n\n(Press q to quit)"
	}
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("Error: " + m.err.Error())
	}

	return fmt.Sprintf(
		"Blueprint generated successfully using %s\n\n%+v\n\n(Press q to quit)",
		m.client.Name(), m.blueprint,
	)
}

// Messages
type blueprintMsg struct{ bp *llm.Blueprint }
type errMsg struct{ err error }