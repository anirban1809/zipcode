package ui

import (
	"fmt"
	"os"
	"strings"
	"zipcode/src/agent"
	"zipcode/src/ui/components"
	"zipcode/src/workspace"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/reflow/wordwrap"
)

type AppModel struct {
	Workspace *workspace.Workspace
	Runtime   *agent.Runtime
	Prompt    textinput.Model
	ViewPort  viewport.Model
	Result    string
	Running   bool
	Tasks     []components.Task
}

func Iniaitalize(workspace *workspace.Workspace) AppModel {

	width, height, _ := getTerminalSize()

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

func getTerminalSize() (int, int, error) {
	fd := (os.Stdout.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
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

	if a.Result != "" {
		b.WriteString(a.Result)
	}

	a.ViewPort.SetContent(renderBottomAligned(b.String(), a.ViewPort.Height))
	a.ViewPort.GotoBottom()
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit
		case "enter":
			a.Running = true
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
			a.Tasks[len(a.Tasks)-1].AppendSub(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" └── "+msg.Message) + "\n")
		} else {
			a.Tasks[len(a.Tasks)-1].Running = false
			a.Tasks[len(a.Tasks)-1].UpdateResult("\n" + lipgloss.NewStyle().Render(msg.Message) + "\n")
		}

		// a.ViewPort.SetContent(renderBottomAligned(a.Result, a.ViewPort.Height))
		// a.ViewPort.GotoBottom()

		a.renderView()
		return a, waitForRuntimeEvent(a.Runtime.GetExecutorEvents())

	case tea.WindowSizeMsg:
		ClearScreen()
		width, height, _ := getTerminalSize()
		a.ViewPort.Width = width - 2
		a.ViewPort.Height = height - 7
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
	banner := lipgloss.NewStyle().Width(a.ViewPort.Width).
		Border(lipgloss.NormalBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		Render(fmt.Sprintf("ZipCode v0.0.1\nCurrent Workspace: %s\n", a.Workspace.RootPath))
	promptView := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderLeft(false).BorderRight(false).Render(a.Prompt.View())
	viewPortView := wordwrap.String(lipgloss.NewStyle().Render(a.ViewPort.View()), a.ViewPort.Width-2)
	return fmt.Sprintf("\n%s\n%s\n%s", banner, viewPortView, promptView)
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
}
