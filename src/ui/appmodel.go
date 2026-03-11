package ui

import (
	"fmt"
	"zipcode/src/agent"
	"zipcode/src/workspace"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	Workspace *workspace.Workspace
	Runtime   *agent.Runtime
	Prompt    textinput.Model
	Result    string
}

func Iniaitalize(workspace *workspace.Workspace) AppModel {
	input := textinput.New()
	input.Placeholder = "Enter text..."
	input.Focus()
	input.CharLimit = 256
	input.Width = 256

	runtime := agent.NewRuntime(workspace)

	return AppModel{
		Workspace: workspace,
		Runtime:   &runtime,
		Prompt:    input,
		Result:    "",
	}
}

func (a AppModel) Init() tea.Cmd {
	return textinput.Blink
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit

		case "enter":
			err := a.Runtime.Run(a.Prompt.Value())
			a.Prompt.SetValue("")

			if err != nil {
				panic(err)
			}
		}

	}

	a.Prompt, cmd = a.Prompt.Update(msg)
	return a, cmd
}

func (a AppModel) View() string {
	banner := "ZipCode v0.0.1"
	v := lipgloss.NewStyle().Render(a.Prompt.View())
	return fmt.Sprintf("%s\n%s", banner, v)
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
}
