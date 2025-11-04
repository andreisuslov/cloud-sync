package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andreisuslov/cloud-sync/internal/ui"
)

var (
	// Version is set via ldflags during build
	Version = "0.1.0-dev"
)

func main() {
	// Initialize the Bubbletea program
	p := tea.NewProgram(
		ui.NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Enable mouse support for scrolling
		tea.WithFPS(30),           // Force more frequent repaints to avoid partial updates
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}