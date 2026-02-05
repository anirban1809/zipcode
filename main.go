package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	header string
	output string
	prompt textinput.Model
}

type RootModel struct {
	model model
}

func initialRootModel() RootModel {
	return RootModel{
		model: initialModel(),
	}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type your command..."
	ti.Focus()

	wd, err := os.Getwd()

	if err != nil {
		wd = ""
	}

	return model{
		header: fmt.Sprintf("ZipCode v0.0.1 | [workspace: %s] | [llm: Claude 3.5]", wd),
		output: "Waiting for input...",
		prompt: ti,
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
	m.model, modelCmd = m.model.Update(msg)
	cmd = append(cmd, modelCmd)
	return m, tea.Batch(cmd...)
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.output = fmt.Sprintf("→ running task: %s", m.prompt.Value())
			m.prompt.SetValue("")
		}
	}

	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

func (m RootModel) View() string {
	content := m.model.View()

	return lipgloss.NewStyle().Border(lipgloss.ASCIIBorder()).Render(content)
}

func (m model) View() string {
	header := lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render(m.header)
	output := lipgloss.NewStyle().Render(m.output)
	prompt := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(80).Render("" + m.prompt.View())
	return fmt.Sprintf("%s\n%s\n%s", header, output, prompt)
}

func main() {
	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
	p := tea.NewProgram(initialRootModel())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		return
	}
}
