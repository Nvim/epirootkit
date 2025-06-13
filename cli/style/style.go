package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderHeight         int     = 16
	ChildModelWidthRatio float32 = 0.6
)

func SelectedTab(tabw, position int) (selected lipgloss.Style) {
	border := lipgloss.NormalBorder()
	switch position {
	case 0:
		border.BottomLeft = "│"
		border.BottomRight = "└"
	case 3:
		border.BottomLeft = "┘"
		border.BottomRight = "│"
	default:
		border.BottomLeft = "┘"
		border.BottomRight = "└"
	}

	border.Bottom = " "
	selected = lipgloss.NewStyle().
		Border(border).
		Width(tabw).
		Height(1).
		Padding(0, 0, 1, 0).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#22AA55"))

	return selected
}

func NormalTab(tabw, position int) (normal lipgloss.Style) {
	// TopLeft:      "┌",
	// TopRight:     "┐",
	bn := lipgloss.Border{
		Top:         " ",
		Bottom:      "─",
		Left:        "",
		Right:       "",
		TopLeft:     "",
		TopRight:    "",
		BottomLeft:  "─",
		BottomRight: "─",
		// BottomRight: "┌",
		// BottomLeft:     "┐",
		MiddleLeft:   "",
		MiddleRight:  "",
		Middle:       "",
		MiddleTop:    "",
		MiddleBottom: "",
	}
	if position == 0 {
		// bn.BottomRight = "┌"
		// bn.BottomLeft =  "┐"
		// bn.BottomLeft =  "@"
		bn.BottomRight = "┐"
	}
	if position == 3 {
		// bn.BottomRight = "@"
		bn.BottomLeft = "┌"
	}
	normal = lipgloss.NewStyle().
		Border(bn, false, false, true, false).
		Padding(1, 0, 0, 0).
		Width(tabw).
		Height(1).
		Align(lipgloss.Center)

	return
}

func MainBoxStyle(tabw, height int) lipgloss.Style {
	normalBorder := lipgloss.Border{
		Top:          " ",
		Bottom:       "─",
		Left:         "│",
		Right:        "│",
		TopLeft:      "│",
		TopRight:     "│",
		BottomLeft:   "└",
		BottomRight:  "┘",
		MiddleLeft:   "├",
		MiddleRight:  "┤",
		Middle:       "┼",
		MiddleTop:    "┬",
		MiddleBottom: "┴",
	}
	return lipgloss.NewStyle().
		Border(normalBorder).
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
