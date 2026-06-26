package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/HappyMonkeyAI/prompt-os/internal/execute"
	"github.com/HappyMonkeyAI/prompt-os/internal/hardware"
	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BlueprintModel handles generation, review, and execution of an LLM blueprint.
type BlueprintModel struct {
	client    llm.LLMClient
	wizard    WizardModel
	hw        hardware.HardwareInfo
	blueprint *llm.Blueprint
	err       error
	loading   bool
	executing bool
	execDone  bool
	execSteps []string
	execErr   error
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
		userPrompt := fmt.Sprintf(
			"User preferences: %v\nHardware: %+v\nGenerate a Linux installation blueprint.",
			m.wizard.Answers(), m.hw,
		)
		fullPrompt := llm.SystemPrompt + "\n\n" + userPrompt

		resp, err := m.client.Generate(context.Background(), fullPrompt)
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
	case execDoneMsg:
		m.executing = false
		m.execDone = true
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			if m.blueprint != nil && !m.executing && !m.execDone {
				return m, m.runConfirm()
			}
		case "d":
			if m.blueprint != nil && !m.executing && !m.execDone {
				m.execSteps, _ = execute.BuildBootstrapPlan(execute.InstallOptions{
					Blueprint: m.blueprint,
					MountRoot: "/mnt/promptos-target",
					DryRun:    true,
				})
				m.execDone = true
			}
		}
	}
	return m, nil
}

func (m BlueprintModel) View() string {
	style := lipgloss.NewStyle().Padding(1, 2)

	if m.loading {
		return style.Render("Generating blueprint with " + m.client.Name() + "...\n\n(Press q to quit)")
	}
	if m.err != nil {
		return style.Render("Error: " + m.err.Error() + "\n\n(Press q to quit)")
	}

	var sb strings.Builder
	sb.WriteString("Blueprint generated with " + m.client.Name() + "\n\n")
	sb.WriteString(fmt.Sprintf("%+v\n\n", m.blueprint))

	if m.executing {
		sb.WriteString("Running...\n")
	} else if m.execDone {
		if m.execErr != nil {
			sb.WriteString("Failed: " + m.execErr.Error() + "\n")
		} else {
			sb.WriteString("Plan:\n")
			for _, s := range m.execSteps {
				sb.WriteString("- " + s + "\n")
			}
		}
	} else {
		sb.WriteString("Actions:\n")
		sb.WriteString("  d  Dry run\n")
		sb.WriteString("  c  Confirm and run\n")
	}

	sb.WriteString("\n(Press q to quit)")
	return style.Render(sb.String())
}

func (m BlueprintModel) runConfirm() tea.Cmd {
	return func() tea.Msg {
		m.executing = true
		_, err := execute.InstallBaseSystem(execute.InstallOptions{
			Blueprint: m.blueprint,
			MountRoot: "/mnt/promptos-target",
			DryRun:    false,
			Confirm:   true,
		}, noopRunner{})
		if err != nil {
			return execDoneMsg{err: err}
		}
		m.execSteps = []string{"base system staged for chroot"}
		return execDoneMsg{}
	}
}

type execDoneMsg struct {
	err error
}
type blueprintMsg struct{ bp *llm.Blueprint }
type errMsg struct{ err error }

type noopRunner struct{}

func (noopRunner) Output(name string, args ...string) ([]byte, error) { return nil, nil }
func (noopRunner) Run(name string, args ...string) error               { return nil }
func (noopRunner) Stat(path string) (os.FileInfo, error)              { return nil, nil }
