package repl

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"jeeves/config"
)

func Run(cfg *config.Config) {
	m := newModel(cfg)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
