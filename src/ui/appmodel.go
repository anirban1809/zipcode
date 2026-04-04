package ui

import (
	"fmt"
	"strings"
	"zipcode/src/agent"
	"zipcode/src/config"
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
	Workspace           *workspace.Workspace
	Runtime             *agent.Runtime
	Prompt              textinput.Model
	ViewPort            viewport.Model
	Result              string
	Tasks               []components.Task
	Conversation        string
	ActiveConversation  string
	Question            components.Question
	FileChangeViewer    components.FileChangeViewer
	CommandsMenu        components.Menu
	Commands            []string
	CommandDescriptions []string
	StatusBar           components.StatusBar
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
		StatusBar:           components.CreateStatusBar(workspace.RootPath, string(config.CurrentModel)),
		ActiveConversation:  "\n",
	}
}

func (a AppModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, a.StatusBar.Init())
}

func waitForRuntimeEvent() tea.Cmd {
	return func() tea.Msg {
		return agent.EventManager.ReadFromChannel(agent.AGENT_OUTPUT_CHANNEL)
	}
}

func waitForFileChangeEvent() tea.Cmd {
	return func() tea.Msg {
		return agent.EventManager.ReadFromChannel(agent.FILE_DIFF_CHANNEL)
	}
}

func (a *AppModel) ProcessQuestion() {
	if a.Question.Question == "" || !a.Question.Selected {
		return
	}

	a.Question.Visible = false
	a.FileChangeViewer.Visible = false
	agent.EventManager.WriteToChannel(agent.AGENT_INPUT_CHANNEL, a.Question.GetSelectedItem())
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

func (a AppModel) GetConversation() string {
	content := a.Conversation + "\n" + a.ActiveConversation
	return wordwrap.String(content, a.ViewPort.Width)
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit
		case "w":
			if a.FileChangeViewer.Visible {
				a.FileChangeViewer.ScrollUp()
				return a, tea.Batch(cmds...)
			}
		case "s":
			if a.FileChangeViewer.Visible {
				a.FileChangeViewer.ScrollDown()
				return a, tea.Batch(cmds...)
			}
		case "enter":
			if strings.HasPrefix(a.Prompt.Value(), "/") {
				a.CommandsMenu.SetVisible(false)
				a.Prompt.SetValue("")
				return a, tea.Batch(tea.Batch(cmds...),
					waitForRuntimeEvent(), waitForFileChangeEvent())
			}

			if a.Prompt.Value() != "" {
				go a.Runtime.Run(a.Prompt.Value())
				a.ActiveConversation += "⏺ " + a.Prompt.Value() + "\n"
				a.StatusBar.SetStatus(components.Status_RUNNING)
				a.ViewPort.SetContent(a.GetConversation())
				a.ViewPort.GotoBottom()
				a.Prompt.SetValue("")
				return a, tea.Batch(tea.Batch(cmds...),
					waitForRuntimeEvent(), waitForFileChangeEvent())
			}

		case "tab":
			if a.StatusBar.GetMode() == components.Mode_PLAN {
				a.StatusBar.SetMode(components.Mode_EDIT)
				return a, tea.Batch(tea.Batch(cmds...),
					waitForRuntimeEvent(), waitForFileChangeEvent())
			}

			a.StatusBar.SetMode(components.Mode_PLAN)
			return a, tea.Batch(tea.Batch(cmds...),
				waitForRuntimeEvent(), waitForFileChangeEvent())

		}

	case agent.ResponseEvent:
		a.StatusBar.UpdateUsage(a.Runtime.InputTokens, a.Runtime.OutputTokens)

		if msg.EventType == agent.Tool {
			a.ActiveConversation += (lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render(" └── "+msg.Message) + "\n")

			a.ViewPort.SetContent(a.GetConversation())
			a.ViewPort.GotoBottom()
			if msg.Question != "" {
				a.SetQuestion(components.CreateQuestion(msg.Question, msg.Options))
				a.Question.Visible = true
			}

		} else {
			a.ActiveConversation += ("\n" + lipgloss.NewStyle().Render(msg.Message) + "\n")
			a.Conversation += "\n" + strings.Replace(a.ActiveConversation, "⏺ ", "✔ ", 1)
			a.ActiveConversation = "\n"
			a.StatusBar.SetStatus(components.Status_IDLE)
			a.ViewPort.SetContent(a.GetConversation())
			a.ViewPort.GotoBottom()
		}
		return a, tea.Batch(waitForRuntimeEvent(), waitForFileChangeEvent())

	case agent.FileChangeEvent:
		changeType := "patch"
		switch msg.ChangeType {
		case agent.FileChange_Create:
			changeType = "create"
		case agent.FileChange_Append:
			changeType = "append"
		}
		a.FileChangeViewer = components.CreateFileChangeViewer(msg.FileName, changeType, msg.Content, msg.Patches)
		return a, tea.Batch(waitForRuntimeEvent(), waitForFileChangeEvent())

	case tea.WindowSizeMsg:
		ClearScreen()
		width, height, _ := utils.GetTerminalSize()
		a.ViewPort.GotoBottom()
		a.ViewPort.Width = width - 2
		a.ViewPort.Height = height - 4
	}

	a.ProcessQuestion()
	cmds = append(cmds, a.ProcessCommandsMenu())
	a.CommandsMenu.SetVisible(strings.HasPrefix(a.Prompt.Value(), "/"))

	var cmd tea.Cmd

	a.Prompt, cmd = a.Prompt.Update(msg)
	cmds = append(cmds, cmd)

	a.Question, cmd = a.Question.Update(msg)
	cmds = append(cmds, cmd)

	a.ViewPort, cmd = a.ViewPort.Update(msg)
	cmds = append(cmds, cmd)

	a.CommandsMenu, cmd = a.CommandsMenu.Update(msg)
	cmds = append(cmds, cmd)

	a.StatusBar, cmd = a.StatusBar.Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a *AppModel) SetQuestion(question components.Question) {
	a.Question = question
}

func (a AppModel) View() string {
	promptView := lipgloss.NewStyle().PaddingBottom(1).Render(a.Prompt.View())
	viewPortView := lipgloss.NewStyle().Padding(1).Render(a.ViewPort.View())
	return fmt.Sprintf("\n%s\n%s%s\n%s%s\n%s", viewPortView, a.FileChangeViewer.View(), a.Question.View(), promptView, a.CommandsMenu.View(), a.StatusBar.View())
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
}
