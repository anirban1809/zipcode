package components

import (
	"fmt"
	"zipcode/src/utils"
	"zipcode/src/workspace"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	Workspace   string
	Model       string
	TokensUsed  int
	TotalTokens int
	ContextUsed string
	Branch      string
	Status      string
}

const (
	Status_IDLE = iota
	Status_RUNNING
	Status_ERROR
)

func CreateStatusBar(workspace string, model string) StatusBar {
	return StatusBar{
		Workspace: workspace,
		Model:     model,
	}
}

func (s StatusBar) Init() tea.Cmd {
	return nil
}

func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	return s, nil
}

func (s StatusBar) View() string {
	width, _, _ := utils.GetTerminalSize()

	topPadding := ""
	bottomPadding := ""
	for range width {
		topPadding += lipgloss.NewStyle().Foreground(lipgloss.Color("#95bdfc")).Render("▄")
		bottomPadding += lipgloss.NewStyle().Foreground(lipgloss.Color("#95bdfc")).Render("▀")
	}

	statusBarView := lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("#95bdfc")).
		Foreground(lipgloss.Color("#000000")).
		Render(fmt.Sprintf("\tIDLE | %s | %s", workspace.AbsToTildePath(s.Workspace), s.Model))
	return fmt.Sprintf("%s\n%s\n%s", topPadding, statusBarView, bottomPadding)
}
