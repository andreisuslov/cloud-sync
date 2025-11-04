package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andreisuslov/cloud-sync/internal/ui"
)

var (
	Version = "0.1.0-dev-debug"
	debugLog *os.File
)

func main() {
	// Open debug log
	var err error
	debugLog, err = os.Create("/tmp/cloud-sync-debug.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create debug log: %v\n", err)
		os.Exit(1)
	}
	defer debugLog.Close()

	logDebug("=== Cloud Sync Debug Session Started ===")
	logDebug("Time: %s", time.Now().Format(time.RFC3339))

	// Wrap the model to log updates
	m := ui.NewModel()
	debugModel := &DebugModel{Model: m}

	p := tea.NewProgram(
		debugModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	logDebug("Program starting...")
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		logDebug("Error: %v", err)
		os.Exit(1)
	}
	logDebug("=== Session Ended ===")
}

type DebugModel struct {
	Model ui.Model
	updateCount int
	lastKey string
}

func (m *DebugModel) Init() tea.Cmd {
	logDebug("Init() called")
	return m.Model.Init()
}

func (m *DebugModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updateCount++
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.lastKey = msg.String()
		logDebug("[Update #%d] KeyMsg: %q (Type: %d)", m.updateCount, msg.String(), msg.Type)
		
		// Log list state before update
		if m.Model.State == 0 { // StateMainMenu
			logDebug("  List Index Before: %d, Total Items: %d", 
				m.Model.List.Index(), len(m.Model.List.Items()))
		}
		
	case tea.WindowSizeMsg:
		logDebug("[Update #%d] WindowSizeMsg: %dx%d", m.updateCount, msg.Width, msg.Height)
		
	case tea.MouseMsg:
		logDebug("[Update #%d] MouseMsg: Type=%d, X=%d, Y=%d", 
			m.updateCount, msg.Type, msg.X, msg.Y)
	}
	
	// Call the actual update
	newModel, cmd := m.Model.Update(msg)
	m.Model = newModel.(ui.Model)
	
	// Log state after update
	if m.lastKey != "" {
		if m.Model.State == 0 { // StateMainMenu
			logDebug("  List Index After: %d", m.Model.List.Index())
		}
		m.lastKey = ""
	}
	
	return m, cmd
}

func (m *DebugModel) View() string {
	view := m.Model.View()
	
	// Log view generation periodically
	if m.updateCount % 10 == 0 {
		logDebug("[View #%d] Generated view, length: %d bytes", m.updateCount, len(view))
	}
	
	return view
}

func logDebug(format string, args ...interface{}) {
	if debugLog != nil {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Fprintf(debugLog, "[%s] ", timestamp)
		fmt.Fprintf(debugLog, format, args...)
		fmt.Fprintln(debugLog)
		debugLog.Sync()
	}
}
