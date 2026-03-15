package components

import (
	"fmt"
	"strings"
	"zipcode/src/config"
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
	Status      Status
}

type Status string

const (
	Status_IDLE    Status = "IDLE"
	Status_RUNNING Status = "RUNNING"
	Status_ERROR   Status = "ERROR"
)

func CreateStatusBar(workspace string, model string) StatusBar {
	return StatusBar{
		Workspace: workspace,
		Model:     model,
		Status:    Status_IDLE,
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

	top := buildPadding(width, "▄")
	bottom := buildPadding(width, "▀")

	bar := renderStatusBar(width, s)

	return fmt.Sprintf("%s\n%s\n%s", top, bar, bottom)
}

func buildPadding(width int, char string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#95bdfc"))

	var b strings.Builder
	for range width {
		b.WriteString(style.Render(char))
	}

	return b.String()
}

func renderStatusBar(width int, s StatusBar) string {
	appVersion := fmt.Sprintf("ZipCode v%s ", config.APP_VERSION)

	status := fmt.Sprintf(
		" %s | Workspace: %s (Branch: main) | Model: %s",
		s.Status,
		workspace.AbsToTildePath(s.Workspace),
		s.Model,
	)

	content := status + utils.FlexGap(width, len(status)+len(appVersion)) + appVersion

	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("#95bdfc")).
		Foreground(lipgloss.Color("#000000")).
		Render(content)
}

func (s *StatusBar) SetStatus(value Status) {
	s.Status = value
}
