package ui

import (
	"fmt"
	"zipcode/src/agent"
	"zipcode/src/workspace"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	Workspace *workspace.Workspace
	Runtime   *agent.Runtime
	Prompt    textinput.Model
	ViewPort  viewport.Model
	Result    string
}

func Iniaitalize(workspace *workspace.Workspace) AppModel {
	input := textinput.New()
	input.Placeholder = "Enter text..."
	input.Focus()
	input.CharLimit = 256
	input.Width = 256
	vp := viewport.New(200, 20)
	vp.SetContent(``)

	runtime := agent.NewRuntime(workspace)

	return AppModel{
		Workspace: workspace,
		Runtime:   &runtime,
		Prompt:    input,
		Result:    "",
		ViewPort:  vp,
	}
}

func (a AppModel) Init() tea.Cmd {
	return textinput.Blink
}

func waitForRuntimeEvent(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit

		case "enter":
			if a.Prompt.Value() != "" {
				go a.Runtime.Run(a.Prompt.Value())
				a.Prompt.SetValue("")
				return a, waitForRuntimeEvent(a.Runtime.GetExecutorEvents())
			}
		}

	case string:
		a.Result += msg + "\n"
		a.ViewPort.SetContent(lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Render(a.Result))
		return a, waitForRuntimeEvent(a.Runtime.GetExecutorEvents())
	}

	a.Prompt, cmd = a.Prompt.Update(msg)
	a.ViewPort, cmd = a.ViewPort.Update(msg)
	return a, cmd
}

func (a AppModel) View() string {
	banner := lipgloss.NewStyle().Width(200).
		Border(lipgloss.NormalBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		Render(fmt.Sprintf("ZipCode v0.0.1\nCurrent Workspace: %s\n", a.Workspace.RootPath))
	promptView := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderLeft(false).BorderRight(false).Render(a.Prompt.View())
	viewPortView := lipgloss.NewStyle().Render(a.ViewPort.View())
	return fmt.Sprintf("\n%s\n%s\n%s", banner, viewPortView, promptView)
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
}
