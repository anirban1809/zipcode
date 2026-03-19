package components

import (
	"fmt"
	"zipcode/src/config"
	"zipcode/src/utils"
	"zipcode/src/workspace"

	"github.com/charmbracelet/bubbles/spinner"
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
	Spinner     spinner.Model
	Mode        Mode
}

type Mode string

const (
	Mode_PLAN Mode = "Plan"
	Mode_EDIT Mode = "Edit"
)

type Status string

const (
	Status_IDLE    Status = "Idle"
	Status_RUNNING Status = "Running"
	Status_ERROR   Status = "Error"
)

func CreateStatusBar(workspace string, model string) StatusBar {
	return StatusBar{
		Workspace: workspace,
		Model:     model,
		Status:    Status_IDLE,
		Mode:      Mode_EDIT,
		Spinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}

func (s StatusBar) Init() tea.Cmd {
	return s.Spinner.Tick
}

func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	var cmd tea.Cmd
	s.Spinner, cmd = s.Spinner.Update(msg)
	return s, cmd
}

func (s StatusBar) View() string {
	width, _, _ := utils.GetTerminalSize()
	bar := renderStatusBar(width, s)

	return fmt.Sprintf("%s", bar)
}

func (s StatusBar) GetMode() Mode {
	return s.Mode
}

func (s *StatusBar) SetMode(mode Mode) {
	s.Mode = mode
}

func renderStatusBar(width int, s StatusBar) string {
	appVersion := fmt.Sprintf("ZipCode v%s ", config.APP_VERSION)

	spinnerView := ""

	if s.Status == Status_RUNNING {
		spinnerView = s.Spinner.View()
	}

	status := fmt.Sprintf(
		" %s%s (%s) | %s (main) | %s",
		spinnerView,
		s.Status,
		s.Mode,
		workspace.AbsToTildePath(s.Workspace),
		s.Model,
	)

	content := status + utils.FlexGap(width, len(status)+len(appVersion)) + appVersion

	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("#144492")).
		Foreground(lipgloss.Color("#d9d9d9")).
		Render(content)
}

func (s *StatusBar) SetStatus(value Status) {
	s.Status = value
}
