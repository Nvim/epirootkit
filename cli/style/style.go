package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderHeight         int     = 16
	ChildModelWidthRatio float32 = 0.6
	LightGreen                   = lipgloss.Color("#11cc11")
	DarkGreen                    = lipgloss.Color("#117211")
	VeryDark                     = lipgloss.Color("#004400")
	Blue                         = lipgloss.Color("#4a8d7e")
	Purple                       = lipgloss.Color("#9472b4")
	Gray                         = lipgloss.Color("#9ea0a2")
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
		Padding(1, 0, 0, 0).
		Align(lipgloss.Center).
		Foreground(LightGreen).
		BorderForeground(VeryDark).
		Bold(true)

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
	if position == 3 {
		// bn.BottomRight = "┌"
		// bn.BottomLeft =  "┐"
		// bn.BottomLeft =  "@"
		bn.BottomRight = "┐"
	}
	if position == 0 {
		// bn.BottomRight = "@"
		bn.BottomLeft = "┌"
	}
	normal = lipgloss.NewStyle().
		Border(bn, false, true, true, true).
		Padding(1, 0, 0, 0).
		Width(tabw - 2).
		Height(1).
		Align(lipgloss.Center).
		BorderForeground(VeryDark)

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
		Height(height).
		BorderForeground(VeryDark).
		Foreground(Gray)
}

func LeftBoxStyle(width, height int) lipgloss.Style {
	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Align(lipgloss.Center).
		Width(width).
		Height(height).
		BorderForeground(VeryDark).
		Foreground(Gray)

	return leftStyle
}
