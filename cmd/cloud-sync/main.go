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
	// Note: Not using tea.WithAltScreen() to allow text selection/copying from terminal
	// Not using tea.WithMouseCellMotion() to allow normal terminal mouse behavior
	p := tea.NewProgram(ui.NewModel())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}