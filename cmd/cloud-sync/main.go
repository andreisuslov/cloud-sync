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
	// Initialize the Bubbletea program with inline rendering
	// This forces full repaints and avoids partial screen updates
	p := tea.NewProgram(
		ui.NewModel(),
		tea.WithMouseCellMotion(), // Enable mouse support for scrolling
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}