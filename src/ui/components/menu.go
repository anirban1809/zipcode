package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Menu struct {
	items            []string
	itemDescriptions []string
	selectedItem     int
	cursor           int
	visible          bool
	selected         bool
}

func CreateMenu(items []string, itemDescriptions []string) Menu {
	return Menu{
		items:            items,
		itemDescriptions: itemDescriptions,
		visible:          false,
	}
}

func (m Menu) IsVisible() bool {
	return m.visible
}

func (m *Menu) SetVisible(value bool) {
	m.visible = value
}

func (m Menu) GetSelectedIndex() int {
	return m.cursor
}

func (m Menu) IsSelected() bool {
	return m.selected
}

func (m Menu) Init() tea.Cmd {
	return nil
}

func (m Menu) Update(msg tea.Msg) (Menu, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor = m.cursor - 1
			}

		case "down":
			if m.cursor < len(m.items)-1 {
				m.cursor = m.cursor + 1
			}

		case "enter":
			m.SetVisible(false)
		}
	}

	return m, nil
}

func (m Menu) View() string {
	if !m.visible {
		return ""
	}
	var menuText strings.Builder
	for i, item := range m.items {
		if m.cursor == i {
			menuText.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#47daff")).Render(fmt.Sprintf("> %-40s %s", item, m.itemDescriptions[i])) + "\n")
			continue
		}
		menuText.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#31a1bd")).Render(fmt.Sprintf("  %-40s %s", item, m.itemDescriptions[i])) + "\n")
	}
	return fmt.Sprintf("\n%s", menuText.String())
}
