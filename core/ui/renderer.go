package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorCyan   = lipgloss.Color("#00FFFF")
	ColorYellow = lipgloss.Color("#FFFF00")
	ColorPurple = lipgloss.Color("#800080")
	ColorGreen  = lipgloss.Color("#00FF00")
	ColorRed    = lipgloss.Color("#FF0000")
	ColorBlue   = lipgloss.Color("#0000FF")
	ColorOrange = lipgloss.Color("#FF8000")
	ColorGray   = lipgloss.Color("#333333")
	ColorDim    = lipgloss.Color("#1A1A1A")
	ColorBg     = lipgloss.Color("#0D0D0D")
	ColorText   = lipgloss.Color("#CCCCCC")
	ColorBorder = lipgloss.Color("#666666")

	BlockStyle = lipgloss.Style{}
	EmptyStyle = lipgloss.NewStyle().Foreground(ColorGray)
	BorderStyle = lipgloss.NewStyle().Foreground(ColorBorder)
)

func GetPieceColor(piece byte) lipgloss.Color {
	switch piece {
	case 'I':
		return ColorCyan
	case 'O':
		return ColorYellow
	case 'T':
		return ColorPurple
	case 'S':
		return ColorGreen
	case 'Z':
		return ColorRed
	case 'J':
		return ColorBlue
	case 'L':
		return ColorOrange
	default:
		return ColorGray
	}
}