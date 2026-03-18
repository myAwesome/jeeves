package repl

import "github.com/charmbracelet/lipgloss"

var (
	colCyan   = lipgloss.Color("6")
	colGreen  = lipgloss.Color("2")
	colRed    = lipgloss.Color("1")
	colGray   = lipgloss.Color("8")

	// Left panel: BorderRight + Padding(0,1)
	// Rendered width = Width(w) + 1 border = w + 1
	panelLeft = lipgloss.NewStyle().
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colGray).
			Padding(0, 1)

	panelLeftFocused = lipgloss.NewStyle().
				BorderRight(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colCyan).
				Padding(0, 1)

	// Right panel: Padding(0,2), no border
	// Rendered width = Width(w)
	panelRight = lipgloss.NewStyle().Padding(0, 2)

	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(colCyan)
	dimStyle    = lipgloss.NewStyle().Foreground(colGray)
	errStyle    = lipgloss.NewStyle().Foreground(colRed)
	greenStyle  = lipgloss.NewStyle().Foreground(colGreen)

	postDateStyle = lipgloss.NewStyle().Bold(true).Foreground(colCyan)
	postIDStyle   = lipgloss.NewStyle().Foreground(colGray)

	loginBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colCyan).
			Padding(1, 4).
			Width(46)
)

// panelLeftW is the Width() value to pass to the left panel style.
// Rendered width of left panel = panelLeftW + 1 (border).
func panelLeftW(total int) int {
	w := total * 30 / 100
	if w < 10 {
		return 10
	}
	return w
}

// panelRightW is the Width() value to pass to the right panel style.
// total = (panelLeftW + 1) + panelRightW  →  panelRightW = total - panelLeftW - 1
func panelRightW(total int) int {
	w := total - panelLeftW(total) - 1
	if w < 10 {
		return 10
	}
	return w
}

// contentLeftW is the usable content width inside the left panel.
// panelLeft has Padding(0,1) = 1 left + 1 right → content = panelLeftW - 2
func contentLeftW(total int) int {
	w := panelLeftW(total) - 2
	if w < 1 {
		return 1
	}
	return w
}

// contentRightW is the usable content width inside the right panel.
// panelRight has Padding(0,2) = 2 left + 2 right → content = panelRightW - 4
func contentRightW(total int) int {
	w := panelRightW(total) - 4
	if w < 1 {
		return 1
	}
	return w
}
