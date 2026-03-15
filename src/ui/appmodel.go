package ui

import (
	"fmt"
	"strings"
	"zipcode/src/agent"
	"zipcode/src/ui/components"
	"zipcode/src/utils"
	"zipcode/src/workspace"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type AppModel struct {
	Workspace *workspace.Workspace
	Runtime   *agent.Runtime
	Prompt    textinput.Model
	ViewPort  viewport.Model
	Result    string
	Tasks     []components.Task
	StatusBar components.StatusBar
}

func Iniaitalize(workspace *workspace.Workspace) AppModel {

	width, height, _ := utils.GetTerminalSize()

	input := textinput.New()
	input.Placeholder = "Enter text..."
	input.Focus()
	input.CharLimit = 256
	input.Width = width - 2
	vp := viewport.New(width-2, height-7)
	vp.SetContent(``)
	runtime := agent.NewRuntime(workspace)

	return AppModel{
		Workspace: workspace,
		Runtime:   &runtime,
		Prompt:    input,
		Result:    "",
		ViewPort:  vp,
		StatusBar: components.CreateStatusBar(workspace.RootPath, "openai/gpt-5.1-codex-mini"),
	}
}

func (a AppModel) Init() tea.Cmd {
	return textinput.Blink
}

func waitForRuntimeEvent(ch <-chan agent.ResponseEvent) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func renderBottomAligned(content string, height int) string {
	lines := strings.Split(content, "\n")
	if len(lines) >= height {
		return content
	}
	padding := height - len(lines)
	return strings.Repeat("\n", padding) + content
}

func (a *AppModel) renderView() {
	var b strings.Builder

	for _, task := range a.Tasks {
		b.WriteString(task.View())
		b.WriteString("\n")
		b.WriteString(task.Subs)
		b.WriteString(task.Result)
		b.WriteString("\n")
	}

	a.ViewPort.SetContent(renderBottomAligned(b.String(), a.ViewPort.Height))
	a.ViewPort.GotoBottom()
}

func (a AppModel) getCurrentTask() *components.Task {
	return &a.Tasks[len(a.Tasks)-1]
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit
		case "enter":
			if a.Prompt.Value() != "" {
				task := components.CreateTask(a.Prompt.Value())
				task.Running = true
				cmds = append(cmds, task.Init())
				a.Tasks = append(a.Tasks, task)
				go a.Runtime.Run(a.Prompt.Value())
				a.renderView()
				a.Prompt.SetValue("")
				return a, tea.Batch(tea.Batch(cmds...),
					waitForRuntimeEvent(a.Runtime.GetExecutorEvents()))
			}
		}

	case agent.ResponseEvent:
		if msg.EventType == agent.Tool {
			a.getCurrentTask().AppendSub(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" └── "+msg.Message) + "\n")
		} else {
			a.getCurrentTask().Running = false
			a.getCurrentTask().UpdateResult("\n" + lipgloss.NewStyle().Render(msg.Message) + "\n")
		}
		a.renderView()
		return a, waitForRuntimeEvent(a.Runtime.GetExecutorEvents())

	case tea.WindowSizeMsg:
		ClearScreen()
		width, height, _ := utils.GetTerminalSize()
		a.ViewPort.Width = width - 2
		a.ViewPort.Height = height - 8
	}

	var cmd tea.Cmd

	a.Prompt, cmd = a.Prompt.Update(msg)
	cmds = append(cmds, cmd)

	for i := range a.Tasks {
		a.Tasks[i], cmd = a.Tasks[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	a.renderView()

	a.ViewPort, cmd = a.ViewPort.Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a AppModel) View() string {
	banner := BannerStyle(fmt.Sprintf("ZipCode v0.0.1\nCurrent Workspace: %s\n", workspace.AbsToTildePath(a.Workspace.RootPath)), a.ViewPort.Width)
	promptView := lipgloss.NewStyle().Render(a.Prompt.View())
	viewPortView := wordwrap.String(lipgloss.NewStyle().Render(a.ViewPort.View()), a.ViewPort.Width-2)
	return fmt.Sprintf("\n%s\n%s\n%s\n%s", banner, viewPortView, promptView, a.StatusBar.View())
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
}
