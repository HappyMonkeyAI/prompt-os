package tui

import (
	"context"
	"fmt"
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

func (m BlueprintModel) generateCmd() tea.Cmd {
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

func (m BlueprintModel) Init() tea.Cmd {
	return m.generateCmd()
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
		m.execErr = msg.err
		if msg.err == nil {
			m.execSteps = msg.steps
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if m.err != nil && !m.loading {
				m.err = nil
				m.loading = true
				return m, m.generateCmd()
			}
		case "c":
			if m.blueprint != nil && !m.executing && !m.execDone {
				m.executing = true
				return m, m.runConfirm()
			}
		case "d":
			if m.blueprint != nil && !m.executing && !m.execDone {
				m.execSteps = nil
				if disk := selectedTargetDisk(m.wizard.Answers()); disk != "" {
					if steps, err := execute.BuildPreparePlan(execute.PrepareOptions{Device: disk, DryRun: true, MountRoot: "/mnt/promptos-target"}); err == nil {
						m.execSteps = append(m.execSteps, steps...)
					} else {
						m.execErr = err
						m.execDone = true
						return m, nil
					}
				}
				steps, err := execute.BuildBootstrapPlan(execute.InstallOptions{
					Blueprint: m.blueprint,
					MountRoot: "/mnt/promptos-target",
					DryRun:    true,
				})
				if err != nil {
					m.execErr = err
					m.execDone = true
					return m, nil
				}
				m.execSteps = append(m.execSteps, steps...)

				if len(m.blueprint.Configs) > 0 {
					configSteps, err := execute.BuildConfigDropPlan(execute.ConfigDropOptions{
						Blueprint: m.blueprint,
						MountRoot: "/mnt/promptos-target",
						DryRun:    true,
					})
					if err != nil {
						m.execErr = err
						m.execDone = true
						return m, nil
					}
					for _, f := range configSteps {
						m.execSteps = append(m.execSteps, fmt.Sprintf("write config: %s", f))
					}
				}
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
		return style.Render("Error: " + m.err.Error() + "\n\n(Press r to retry · Press q to quit)")
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
		var runSteps []string

		if disk := selectedTargetDisk(m.wizard.Answers()); disk != "" {
			prepRes, err := execute.PrepareDisk(execute.PrepareOptions{Device: disk, ConfirmWipe: true, MountRoot: "/mnt/promptos-target"}, execute.DefaultRunner)
			if err != nil {
				return execDoneMsg{err: err}
			}
			runSteps = append(runSteps, prepRes.Steps...)
		}

		instRes, err := execute.InstallBaseSystem(execute.InstallOptions{
			Blueprint: m.blueprint,
			MountRoot: "/mnt/promptos-target",
			DryRun:    false,
			Confirm:   true,
		}, execute.DefaultRunner)
		if err != nil {
			return execDoneMsg{err: err}
		}
		runSteps = append(runSteps, instRes.Steps...)

		if len(m.blueprint.Configs) > 0 {
			dropRes, err := execute.ApplyConfigDrop(execute.ConfigDropOptions{
				Blueprint: m.blueprint,
				MountRoot: "/mnt/promptos-target",
				DryRun:    false,
				Confirm:   true,
			})
			if err != nil {
				return execDoneMsg{err: err}
			}
			for _, f := range dropRes.Files {
				runSteps = append(runSteps, fmt.Sprintf("write config: %s", f))
			}
		}

		return execDoneMsg{steps: runSteps}
	}
}

type execDoneMsg struct {
	steps []string
	err   error
}
type blueprintMsg struct{ bp *llm.Blueprint }
type errMsg struct{ err error }

func selectedTargetDisk(answers map[string]string) string {
	disk := strings.TrimSpace(answers["target_disk"])
	if disk == "" || disk == "skip disk prep" {
		return ""
	}
	return disk
}
