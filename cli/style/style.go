package style

import "github.com/charmbracelet/lipgloss"

var (
	HeaderHeight         int     = 16
	ChildModelWidthRatio float32 = 0.6
)

func TabStyles(tabw int) (selected lipgloss.Style, normal lipgloss.Style) {
	selected = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, false, true).
		Width(tabw).
		Height(1).
		Padding(0, 0, 1, 0).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#22AA55"))

	normal = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Padding(1, 0, 0, 0).
		Inherit(selected)

	return
}

func MainBoxStyle(tabw, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).UnsetBorderTop().
		Padding(2).
		Width(tabw * 4).
		Height(height)
}

func LeftBoxStyle(width, height int) lipgloss.Style {
	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Align(lipgloss.Center).
		Width(width).
		Height(height)

	return leftStyle
}
