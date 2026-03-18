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
)

type AppModel struct {
	Workspace           *workspace.Workspace
	Runtime             *agent.Runtime
	Prompt              textinput.Model
	ViewPort            viewport.Model
	Result              string
	Tasks               []components.Task
	Conversation        string
	ActiveConversation  string
	Question            components.Question
	CommandsMenu        components.Menu
	Commands            []string
	CommandDescriptions []string
}

func Iniaitalize(workspace *workspace.Workspace) AppModel {
	width, height, _ := utils.GetTerminalSize()

	input := textinput.New()
	input.Focus()
	input.CharLimit = 1024
	input.Width = width - 2
	vp := viewport.New(width-2, height-4)
	vp.SetContent(``)
	runtime := agent.NewRuntime(workspace)

	items := []string{"/models", "/help", "/exit"}
	itemDescriptions := []string{"Select model", "Get help", "Exit ZipCode"}

	return AppModel{
		Workspace:           workspace,
		Runtime:             &runtime,
		Prompt:              input,
		ViewPort:            vp,
		CommandsMenu:        components.CreateMenu(items, itemDescriptions),
		Commands:            items,
		CommandDescriptions: itemDescriptions,
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

func (a AppModel) getCurrentTask() *components.Task {
	return &a.Tasks[len(a.Tasks)-1]
}

func (a *AppModel) ProcessQuestion() {
	if a.Question.Question == "" || !a.Question.Selected {
		return
	}

	a.Question.Visible = false
	a.Runtime.Executor.MessageChannel <- a.Question.GetSelectedItem()
	a.Question.Selected = false
}

func (a *AppModel) ProcessCommandsMenu() tea.Cmd {
	if !a.CommandsMenu.IsVisible() || !a.CommandsMenu.IsSelected() {
		return nil
	}
	a.CommandsMenu.SetVisible(false)
	i := a.CommandsMenu.GetSelectedIndex()

	switch a.Commands[i] {
	case "/model":
		break
	case "/help":
		break
	case "/exit":
		return tea.Quit
	}

	return nil
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit
		case "enter":
			if strings.HasPrefix(a.Prompt.Value(), "/") {
				a.CommandsMenu.SetVisible(false)
				a.Prompt.SetValue("")
				return a, tea.Batch(tea.Batch(cmds...),
					waitForRuntimeEvent(a.Runtime.GetExecutorEventChannel()))
			}

			if a.Prompt.Value() != "" {
				task := components.CreateTask(a.Prompt.Value())
				task.Running = true
				cmds = append(cmds, task.Init())
				a.Tasks = append(a.Tasks, task)
				go a.Runtime.Run(a.Prompt.Value())
				a.ActiveConversation += a.Prompt.Value() + "\n"
				a.Prompt.SetValue("")
				return a, tea.Batch(tea.Batch(cmds...),
					waitForRuntimeEvent(a.Runtime.GetExecutorEventChannel()))
			}

		}

	case agent.ResponseEvent:
		if msg.EventType == agent.Tool {
			a.ActiveConversation += (lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" └── "+msg.Message) + "\n")
			if msg.Question != "" {
				a.SetQuestion(components.CreateQuestion(msg.Question, msg.Options))
				a.Question.Visible = true
			}

		} else {
			a.getCurrentTask().Running = false
			a.ActiveConversation += ("\n" + lipgloss.NewStyle().Render(msg.Message) + "\n")
		}
		return a, waitForRuntimeEvent(a.Runtime.GetExecutorEventChannel())

	case tea.WindowSizeMsg:
		ClearScreen()
		width, _, _ := utils.GetTerminalSize()
		a.ViewPort.Width = width - 2
	}

	a.ProcessQuestion()
	cmds = append(cmds, a.ProcessCommandsMenu())
	a.CommandsMenu.SetVisible(strings.HasPrefix(a.Prompt.Value(), "/"))

	var cmd tea.Cmd

	a.Prompt, cmd = a.Prompt.Update(msg)
	cmds = append(cmds, cmd)

	for i := range a.Tasks {
		a.Tasks[i], cmd = a.Tasks[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	// a.renderView()

	a.Question, cmd = a.Question.Update(msg)
	cmds = append(cmds, cmd)

	a.ViewPort, cmd = a.ViewPort.Update(msg)
	cmds = append(cmds, cmd)

	a.CommandsMenu, cmd = a.CommandsMenu.Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a *AppModel) SetQuestion(question components.Question) {
	a.Question = question
}

func (a AppModel) View() string {
	promptView := lipgloss.NewStyle().Render(a.Prompt.View())
	// viewPortView := a.ViewPort.View()
	return fmt.Sprintf("\n%s\n%s\n%s%s", a.ActiveConversation, a.Question.View(), promptView, a.CommandsMenu.View())
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
}
