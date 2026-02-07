package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type model struct {
	header      string
	promptWidth int
	prompt      textinput.Model
	result      viewport.Model
}

type RootModel struct {
	model  model
	width  int
	height int
}

func initialRootModel(width int, height int) RootModel {
	return RootModel{
		model: initialModel(width - 7),
		width: width, height: height,
	}
}

func initialModel(promptWidth int) model {
	ti := textinput.New()
	ti.Placeholder = "Start here...."
	ti.Focus()

	wd, err := os.Getwd()

	if err != nil {
		wd = ""
	}

	vp := viewport.New(100, 20)
	vp.SetContent("Thinking...")

	return model{
		header:      fmt.Sprintf("ZipCode v0.0.1 \nWorkspace: %s \nModel: claude-4-sonnet", wd),
		prompt:      ti,
		promptWidth: promptWidth,
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.model.Init()
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd []tea.Cmd
	var modelCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.model.promptWidth = msg.Width - 6
		m.model.result.SetContent("Thinking...")
		m.model.result.Width = msg.Width - 8
		m.model.result.Height = msg.Height - 11
	}

	m.model, modelCmd = m.model.Update(msg)

	cmd = append(cmd, modelCmd)

	return m, tea.Batch(cmd...)
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			m.prompt.SetValue("")
		}
	}

	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

func (m RootModel) View() string {
	content := m.model.View()
	return lipgloss.NewStyle().Height(m.height - 2).Width(m.width - 2).Border(lipgloss.RoundedBorder()).PaddingLeft(1).PaddingRight(1).Render(content)
}

func (m model) View() string {
	header := lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render(m.header)
	prompt := lipgloss.NewStyle().Border(lipgloss.ASCIIBorder()).Width(m.promptWidth).PaddingBottom(1).Render("" + m.prompt.View())
	result := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Render(m.result.View())
	return fmt.Sprintf("%s\n%s\n%s", header, prompt, result)
}

func main() {

	fd := int(os.Stdout.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		fmt.Println("Failed to get terminal size. Exiting")
	}
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
	p := tea.NewProgram(initialRootModel(width, height))

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		return
	}
}
