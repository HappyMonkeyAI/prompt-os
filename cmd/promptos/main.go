package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/HappyMonkeyAI/prompt-os/internal/hardware"
	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
	"github.com/HappyMonkeyAI/prompt-os/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// appStage tracks which screen is currently active.
type appStage int

const (
	stageSplash    appStage = iota // welcome screen
	stageProvider                  // provider + API key selection
	stageWizard                    // OS preference questions
	stageBlueprint                 // LLM call + review + dry-run
)

// AppModel is the root Bubble Tea model. It owns a stage and delegates
// rendering / input to whichever sub-model is active.
type AppModel struct {
	stage     appStage
	hw        hardware.HardwareInfo
	provider  tui.ProviderModel
	wizard    tui.WizardModel
	blueprint tui.BlueprintModel

	// populated after provider selection
	llmClient llm.LLMClient

	// error to display if LLM client construction fails
	initErr error

	width  int
	height int
}

func newAppModel() AppModel {
	return AppModel{
		stage:    stageSplash,
		hw:       hardware.Scan(),
		provider: tui.NewProviderModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.provider.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Window resize — pass through to all stages
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	// Provider confirmed → build LLM client → advance to wizard
	case tui.ProviderDoneMsg:
		client, err := llm.NewClientFromProvider(msg.Provider, msg.APIKey, msg.BaseURL, msg.Model)
		if err != nil {
			m.initErr = err
			return m, nil
		}
		m.llmClient = client
		m.wizard = tui.NewWizardModel()
		m.stage = stageWizard
		return m, m.wizard.Init()

	// Wizard done → build blueprint model → advance to blueprint
	case tui.WizardDoneMsg:
		m.blueprint = tui.NewBlueprintModel(m.llmClient, m.wizard, m.hw)
		m.stage = stageBlueprint
		return m, m.blueprint.Init()
	}

	// Route input to the active stage
	switch m.stage {

	case stageSplash:
		// Any keypress advances past the splash
		if _, ok := msg.(tea.KeyMsg); ok {
			km := msg.(tea.KeyMsg)
			if km.String() == "ctrl+c" || km.String() == "q" {
				return m, tea.Quit
			}
			m.stage = stageProvider
			return m, m.provider.Init()
		}

	case stageProvider:
		updated, cmd := m.provider.Update(msg)
		m.provider = updated.(tui.ProviderModel)
		return m, cmd

	case stageWizard:
		updated, cmd := m.wizard.Update(msg)
		m.wizard = updated.(tui.WizardModel)
		return m, cmd

	case stageBlueprint:
		updated, cmd := m.blueprint.Update(msg)
		m.blueprint = updated.(tui.BlueprintModel)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	muted  := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	if m.initErr != nil {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF4466")).
			Padding(1, 3).
			Render("Error: " + m.initErr.Error() + "\n\nPress q to quit")
	}

	switch m.stage {
	case stageSplash:
		return renderSplash(accent, muted)
	case stageProvider:
		return m.provider.View()
	case stageWizard:
		return m.wizard.View()
	case stageBlueprint:
		return m.blueprint.View()
	}
	return ""
}

// renderSplash draws the welcome screen.
func renderSplash(accent, muted lipgloss.Style) string {
	logo := strings.Join([]string{
		`  ____                            _   ___  ____`,
		` |  _ \ _ __ ___  _ __ ___  _ __| |_/ _ \/ ___|`,
		` | |_) | '__/ _ \| '_ ` + "`" + ` _ \| '_ \| | | | \___ \`,
		` |  __/| | | (_) | | | | | | |_) | | |_| |___) |`,
		` |_|   |_|  \___/|_| |_| |_| .__/|_|\___/|____/`,
		`                            |_|`,
	}, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 4)

	content := accent.Render(logo) + "\n\n" +
		"  AI-powered Linux installer\n\n" +
		muted.Render("  Press any key to begin · q to quit")

	return box.Render(content)
}

func main() {
	p := tea.NewProgram(
		newAppModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "promptos: %v\n", err)
		os.Exit(1)
	}
}
