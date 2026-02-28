package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"zipcode/src/agent"
	"zipcode/src/workspace"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	header        string
	workspace     string
	currentModel  string
	promptWidth   int
	prompt        textinput.Model
	result        viewport.Model
	commands      []string
	commandDesc   map[string]string
	modelOptions  []string
	filtered      []string
	selected      int
	menuOffset    int
	showMenu      bool
	menuKind      string
	waiting       bool
	spinnerIndex  int
	pendingPrompt string
	clearOnExit   bool
	spinner       spinner.Model
	spinning      bool
}

type RootModel struct {
	model  model
	width  int
	height int
}

const maxCommandMenuRows = 6
const headerExtraHeight = 1
const footerHeight = 2

var spinnerFrames = spinner.Spinner{
	Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	FPS:    time.Second / 10, //nolint:mnd
}

const spinnerInterval = 120 * time.Millisecond
const mockResponseDelay = 3 * time.Second

type spinnerTickMsg struct{}
type spinnerDoneMsg struct{}

func (m *RootModel) adjustLayout() {
	m.model.promptWidth = m.width - 6
	m.model.result.Width = m.width - 8

	baseHeight := m.height - 11 - headerExtraHeight - footerHeight
	if m.model.showMenu && len(m.model.filtered) > 0 {
		// menu rows + spacer line
		baseHeight -= m.model.visibleMenuCount() + 1
	}
	if baseHeight < 1 {
		baseHeight = 1
	}
	m.model.result.Height = baseHeight
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

	initialContent := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(initialResultContent())

	vp.SetContent(initialContent)

	m := model{
		workspace:    wd,
		currentModel: "claude-4-sonnet",
		prompt:       ti,
		promptWidth:  promptWidth,
		result:       vp,
		commands: []string{
			"/init", "/models", "/quit", "/help", "/status",
		},
		commandDesc: map[string]string{
			"/init":   "Initialize the current workspace session",
			"/models": "Open model picker and switch active model",
			"/quit":   "Exit the application",
			"/help":   "Show help and available command usage",
			"/status": "Display current session and environment status",
		},
		modelOptions: []string{"claude-4-sonnet", "claude-4-opus", "gemini-2.0-flash", "gpt-4.1"},
		selected:     0,
		spinner:      spinner.New(spinner.WithSpinner(spinnerFrames)),
	}
	m.updateHeader()
	return m
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
	}

	m.model, modelCmd = m.model.Update(msg)
	m.adjustLayout()

	cmd = append(cmd, modelCmd)

	return m, tea.Batch(cmd...)
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case spinner.TickMsg:
		if !m.spinning {
			return m, nil
		}

		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case planningDoneMsg:
		if !m.waiting {
			return m, nil
		}
		promptData := m.pendingPrompt
		// m.waiting = false
		m.spinnerIndex = 0
		m.result.SetContent(fmt.Sprintf("Thinking: %s", promptData))
		return m, thinkCmd()

	case thinkingDoneMsg:
		if !m.waiting {
			return m, nil
		}
		promptData := m.pendingPrompt
		m.waiting = false
		m.spinnerIndex = 0
		m.pendingPrompt = ""
		m.result.SetContent(fmt.Sprintf("Running: %s", promptData))
		return m, nil

	case tea.KeyMsg:
		if m.waiting {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
				fmt.Print("\nUsage Summary\r\n---------------------------\r\nTotal Tokens Uploaded: 3000\r\nTotal Tokens Downloaded: 3345\n\n")
				return m, tea.Quit
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
			fmt.Print("\nUsage Summary\r\n---------------------------\r\nTotal Tokens Uploaded: 3000\r\nTotal Tokens Downloaded: 3345\n\n")
			return m, tea.Quit

		case "shift+enter":
			promptData := m.prompt.Value()
			promptData = fmt.Sprintf("%s\n", promptData)

			m.prompt.SetValue(promptData)

		case "up":
			if m.showMenu && len(m.filtered) > 0 {
				m.selected--
				if m.selected < 0 {
					m.selected = len(m.filtered) - 1
				}
				m.ensureMenuSelectionVisible()
				return m, nil
			}
			m.result.ScrollUp(1)
			return m, nil

		case "down":
			if m.showMenu && len(m.filtered) > 0 {
				m.selected = (m.selected + 1) % len(m.filtered)
				m.ensureMenuSelectionVisible()
				return m, nil
			}
			m.result.ScrollUp(1)
			return m, nil

		case "tab":
			if m.showMenu && len(m.filtered) > 0 {
				if m.menuKind == "models" {
					return m.selectModel(m.filtered[m.selected])
				}
				return m.runCommand(m.filtered[m.selected])
			}

		case "enter":
			promptData := strings.TrimSpace(m.prompt.Value())

			if m.showMenu && len(m.filtered) > 0 {
				if m.menuKind == "models" {
					return m.selectModel(m.filtered[m.selected])
				}
				return m.runCommand(m.filtered[m.selected])
			}

			m.prompt.SetValue("")
			m.showMenu = false
			m.filtered = nil
			m.selected = 0
			m.menuOffset = 0

			if strings.HasPrefix(promptData, "/") {
				return m.runCommand(promptData)
			}

			return m.startMockResponse(promptData)
		}
	}

	m.prompt, cmd = m.prompt.Update(msg)
	m.updateCommandSuggestions()
	return m, cmd
}

func (m RootModel) View() string {
	content := m.model.View()
	return lipgloss.NewStyle().
		Height(m.height - 2).
		Width(m.width - 2).
		Foreground(lipgloss.Color("252")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#036ee8")).
		PaddingLeft(1).
		PaddingRight(1).
		Render(content)
}

// main view function
func (m model) View() string {

	var spinnerContent string
	if m.spinning {
		spinnerContent = m.spinner.View()
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4b83fb")).Render("ZIPCODE: v0.0.1")
	headerInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("36")).
		Render(m.header)
	header := lipgloss.NewStyle().
		Render(lipgloss.JoinVertical(lipgloss.Left, title, headerInfo))
	prompt := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false).BorderForeground(lipgloss.Color("245")).Width(m.promptWidth).Render("" + m.prompt.View())
	result := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Render(fmt.Sprintf("%s %s", spinnerContent, m.result.View()))
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(footerText())

	menu := ""
	if m.showMenu && len(m.filtered) > 0 {
		var rows []string
		start, end := m.menuWindow()
		maxCommandWidth := 0
		if m.menuKind == "commands" {
			for _, command := range m.filtered[start:end] {
				if len(command) > maxCommandWidth {
					maxCommandWidth = len(command)
				}
			}
		}
		for i, command := range m.filtered[start:end] {
			index := start + i
			prefix := "  "
			commandStyle := lipgloss.NewStyle()
			descriptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
			if index == m.selected {
				prefix = "-> "
				commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
				descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			}

			if m.menuKind == "commands" {
				desc := m.commandDesc[command]
				if desc != "" {
					rows = append(rows, commandStyle.Render(fmt.Sprintf("%s%-*s", prefix, maxCommandWidth+20, command))+descriptionStyle.Render(fmt.Sprintf("  %s", desc)))
					continue
				}
			}

			rows = append(rows, commandStyle.Render(prefix+command))
		}
		menu = lipgloss.NewStyle().Width(m.promptWidth).Render(strings.Join(rows, "\n"))
	}

	if menu != "" {
		return fmt.Sprintf("%s\n%s\n%s\n\n%s\n\n\n%s", header, prompt, menu, result, footer)
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n\n\n\n%s", header, prompt, result, footer)
}

func (m *model) updateCommandSuggestions() {
	if m.menuKind == "models" {
		return
	}

	value := strings.TrimSpace(m.prompt.Value())
	if !strings.HasPrefix(value, "/") {
		m.showMenu = false
		m.filtered = nil
		m.selected = 0
		m.menuOffset = 0
		m.menuKind = ""
		return
	}

	m.filtered = m.filtered[:0]
	for _, command := range m.commands {
		if strings.HasPrefix(command, value) {
			m.filtered = append(m.filtered, command)
		}
	}

	m.showMenu = len(m.filtered) > 0
	m.menuKind = "commands"
	if !m.showMenu {
		m.selected = 0
		m.menuOffset = 0
		m.menuKind = ""
		return
	}

	if m.selected >= len(m.filtered) {
		m.selected = 0
	}
	m.ensureMenuSelectionVisible()
}

func (m model) runCommand(command string) (model, tea.Cmd) {
	m.prompt.SetValue("")
	m.showMenu = false
	m.filtered = nil
	m.selected = 0
	m.menuOffset = 0
	m.menuKind = ""

	switch command {
	case "/init":
		m.result.SetContent("Initialized workspace")
		return m, nil
	case "/models":
		m.showMenu = true
		m.menuKind = "models"
		m.filtered = append([]string(nil), m.modelOptions...)
		m.selected = 0
		m.menuOffset = 0
		m.result.SetContent("Select a model")
		return m, nil
	case "/quit":
		m.clearOnExit = true
		return m, tea.Quit
	default:
		if strings.HasPrefix(command, "/") {
			m.result.SetContent(fmt.Sprintf("Executed mock command: %s", command))
			return m, nil
		}
		m.result.SetContent(fmt.Sprintf("Unknown command: %s", command))
		return m, nil
	}
}

func (m *model) updateHeader() {
	m.header = fmt.Sprintf(
		"Workspace: %s\nModel: %s",
		m.workspace,
		m.currentModel,
	)
}

func (m model) selectModel(name string) (model, tea.Cmd) {
	m.currentModel = name
	m.updateHeader()
	m.showMenu = false
	m.menuKind = ""
	m.filtered = nil
	m.selected = 0
	m.menuOffset = 0
	m.result.SetContent(fmt.Sprintf("Selected model: %s", name))
	return m, nil
}

func (m model) visibleMenuCount() int {
	if !m.showMenu || len(m.filtered) == 0 {
		return 0
	}
	if m.menuKind == "commands" && len(m.filtered) > maxCommandMenuRows {
		return maxCommandMenuRows
	}
	return len(m.filtered)
}

func (m model) menuWindow() (int, int) {
	if m.menuKind != "commands" || len(m.filtered) <= maxCommandMenuRows {
		return 0, len(m.filtered)
	}
	start := m.menuOffset
	if start < 0 {
		start = 0
	}
	maxStart := len(m.filtered) - maxCommandMenuRows
	if start > maxStart {
		start = maxStart
	}
	return start, start + maxCommandMenuRows
}

func (m *model) ensureMenuSelectionVisible() {
	if m.menuKind != "commands" || len(m.filtered) <= maxCommandMenuRows {
		m.menuOffset = 0
		return
	}
	if m.selected < m.menuOffset {
		m.menuOffset = m.selected
		return
	}
	end := m.menuOffset + maxCommandMenuRows
	if m.selected >= end {
		m.menuOffset = m.selected - maxCommandMenuRows + 1
	}
}

func (m model) startMockResponse(promptData string) (model, tea.Cmd) {
	m.waiting = true
	m.spinnerIndex = 0
	m.pendingPrompt = promptData

	if !m.waiting {
		return m, nil
	}

	styledPrompt := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(m.pendingPrompt)
	m.result.SetContent(fmt.Sprintf("Planning...\n\n> %s", styledPrompt))
	m.spinning = true
	return m, tea.Batch(m.spinner.Tick, planCmd())
}

type planningDoneMsg struct{}
type thinkingDoneMsg struct{}

func planCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(600 * time.Millisecond)
		return planningDoneMsg{}
	}
}

func thinkCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(600 * time.Millisecond)
		return thinkingDoneMsg{}
	}
}

func initialResultContent() string {
	lines := []string{
		"Getting Started",
		"------------------",
		"1. Type a prompt and press Enter to run it.",
		"2. Type / to open command suggestions.",
		"3. Use Up/Down to navigate menus.",
		"4. Press Enter to execute a selected command.",
		"5. Run /models to switch the active model.",
		"6. Press Ctrl+C or Esc to quit.",
	}

	return strings.Join(lines, "\n")
}

func footerText() string {
	return strings.Join([]string{
		"Shortcuts: ⏎ submit | ⇧+↵ newline | ↑/↓ menu-or-scroll | ⇥ select | ⌃+C quit",
		"Session (mock): Uptime 00:12:47 | Input 3.2k tok | Output 4.8k tok | Latency 420ms",
	}, "\n")
}

func main() {
	// fd := int(os.Stdout.Fd())
	// width, height, err := term.GetSize(fd)
	// if err != nil {
	// 	fmt.Println("Failed to get terminal size. Exiting")
	// }
	// fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
	// p := tea.NewProgram(initialRootModel(width, height))

	// finalModel, err := p.Run()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// if root, ok := finalModel.(RootModel); ok && root.model.clearOnExit {
	// 	fmt.Print("\033[H\033[2J\033[3J") /* clear the screen */
	// }

	workspace := workspace.Load("/Users/anirban/Documents/Code/zipcode")
	runtime := agent.NewRuntime(&workspace)
	runtime.Run("Build this for me")
}
