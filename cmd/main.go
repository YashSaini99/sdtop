package main

import (
	"fmt"
	"os"

	"sdtop/internal/systemd"
	"sdtop/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create systemd manager
	manager, err := systemd.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to systemd: %v\n", err)
		os.Exit(1)
	}
	defer manager.Close()

	// Create log reader
	logReader, err := systemd.NewLogReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log reader: %v\n", err)
		os.Exit(1)
	}
	defer logReader.Close()

	// Create UI model
	model, err := ui.NewModel(manager, logReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create UI: %v\n", err)
		os.Exit(1)
	}

	// Create Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
