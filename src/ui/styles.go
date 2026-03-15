package ui

import "github.com/charmbracelet/lipgloss"

func BannerStyle(value string, width int) string {
	return lipgloss.NewStyle().Width(width).
		Border(lipgloss.NormalBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).Render(value)
}
