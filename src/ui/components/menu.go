package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
	return m.selectedItem
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
				m.cursor--
			}

		case "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter":
			m.selectedItem = m.cursor
			m.selected = true
		}
	}

	return m, nil
}

func (m Menu) View() string {
	if !m.visible {
		return ""
	}

	menuText := ""

	for i, item := range m.items {
		if m.cursor == i {
			menuText += fmt.Sprintf(" > %s\t%s\n", item, m.itemDescriptions[i])
			continue
		}

		menuText += fmt.Sprintf("   %s\t%s\n", item, m.itemDescriptions[i])
	}

	return fmt.Sprintf("\n%s", menuText)
}
