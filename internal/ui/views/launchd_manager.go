package views

import (
	"fmt"
	"strings"

	"github.com/andreisuslov/cloud-sync/internal/launchd"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// LaunchdAction represents an action that can be performed on a LaunchAgent
type LaunchdAction int

const (
	LaunchdActionNone LaunchdAction = iota
	LaunchdActionLoad
	LaunchdActionUnload
	LaunchdActionStart
	LaunchdActionStop
	LaunchdActionRemove
)

// LaunchdManagerModel represents the LaunchAgent manager model
type LaunchdManagerModel struct {
	launchdManager *launchd.Manager
	status         *launchd.Status
	width          int
	height         int
	err            error
	processing     bool
	message        string
	selectedAction int
	actions        []string
}

// NewLaunchdManagerModel creates a new LaunchAgent manager model
func NewLaunchdManagerModel(launchdManager *launchd.Manager, width, height int) LaunchdManagerModel {
	return LaunchdManagerModel{
		launchdManager: launchdManager,
		width:          width,
		height:         height,
		selectedAction: 0,
		actions: []string{
			"Load Agent",
			"Unload Agent",
			"Start Manually",
			"Stop Agent",
			"Remove Agent",
			"Refresh Status",
		},
	}
}

// Init implements tea.Model
func (m LaunchdManagerModel) Init() tea.Cmd {
	return m.refreshStatus()
}

// Update implements tea.Model
func (m LaunchdManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.processing {
			return m, nil // Ignore input while processing
		}

		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.selectedAction > 0 {
				m.selectedAction--
			}

		case "down", "j":
			if m.selectedAction < len(m.actions)-1 {
				m.selectedAction++
			}

		case "enter":
			return m, m.executeAction()

		case "r":
			return m, m.refreshStatus()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case *launchd.Status:
		m.status = msg
		m.processing = false
		m.err = nil
		return m, nil

	case ActionResult:
		m.processing = false
		if msg.Error != nil {
			m.err = msg.Error
			m.message = ""
		} else {
			m.message = msg.Message
			m.err = nil
		}
		return m, m.refreshStatus()

	case error:
		m.processing = false
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m LaunchdManagerModel) View() string {
	var b strings.Builder

	helper := NewViewHelper(m.width, m.height)

	// Header
	title := "LaunchAgent Manager"
	subtitle := fmt.Sprintf("Managing: %s", m.launchdManager.GetLabel())
	b.WriteString(helper.RenderHeader(title, subtitle))

	// Error or success message
	if m.err != nil {
		b.WriteString(styles.RenderError("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	} else if m.message != "" {
		b.WriteString(styles.RenderSuccess(m.message))
		b.WriteString("\n\n")
	}

	// Status display
	if m.processing {
		b.WriteString(styles.RenderInfo("Processing..."))
		b.WriteString("\n")
	} else if m.status != nil {
		b.WriteString(m.renderStatus())
	} else {
		b.WriteString(styles.RenderInfo("Loading status..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Actions menu
	b.WriteString(styles.RenderSubtitle("Actions:"))
	b.WriteString("\n\n")

	for i, action := range m.actions {
		cursor := "  "
		if i == m.selectedAction {
			cursor = styles.RenderHighlight("> ")
		}

		// Disable certain actions based on status
		disabled := false
		if m.status != nil {
			switch i {
			case 0: // Load
				disabled = m.status.Loaded
			case 1: // Unload
				disabled = !m.status.Loaded
			case 2: // Start
				disabled = !m.status.Loaded || m.status.Running
			case 3: // Stop
				disabled = !m.status.Running
			case 4: // Remove
				// Can always remove
			}
		}

		actionText := action
		if disabled {
			actionText = styles.RenderMuted(action + " (disabled)")
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, actionText))
	}

	// Footer
	helpText := "↑/↓: Navigate • enter: Execute • r: Refresh • q/esc: Back"
	b.WriteString(helper.RenderFooter(helpText))

	return b.String()
}

// renderStatus renders the LaunchAgent status
func (m LaunchdManagerModel) renderStatus() string {
	var b strings.Builder

	b.WriteString(styles.RenderSubtitle("Current Status:"))
	b.WriteString("\n\n")

	// Loaded status
	loadedIcon := "✗"
	loadedColor := styles.RenderError
	loadedText := "Not Loaded"
	if m.status.Loaded {
		loadedIcon = "✓"
		loadedColor = styles.RenderSuccess
		loadedText = "Loaded"
	}
	b.WriteString(fmt.Sprintf("  Loaded:     %s %s\n", loadedIcon, loadedColor(loadedText)))

	// Running status
	runningIcon := "✗"
	runningColor := styles.RenderMuted
	runningText := "Not Running"
	if m.status.Running {
		runningIcon = "✓"
		runningColor = styles.RenderSuccess
		runningText = "Running"
	}
	b.WriteString(fmt.Sprintf("  Running:    %s %s\n", runningIcon, runningColor(runningText)))

	// PID
	if m.status.Running && m.status.PID > 0 {
		b.WriteString(fmt.Sprintf("  PID:        %s\n", styles.RenderInfo(fmt.Sprintf("%d", m.status.PID))))
	}

	// Last exit status
	if m.status.LastExit != 0 {
		b.WriteString(fmt.Sprintf("  Last Exit:  %s\n", 
			styles.RenderWarning(fmt.Sprintf("%d (error)", m.status.LastExit))))
	} else if m.status.Loaded {
		b.WriteString(fmt.Sprintf("  Last Exit:  %s\n", 
			styles.RenderSuccess("0 (success)")))
	}

	// Plist file location
	b.WriteString(fmt.Sprintf("\n  Plist:      %s\n", 
		styles.RenderMuted(m.launchdManager.GetPlistPath())))

	return b.String()
}

// refreshStatus returns a command to refresh the LaunchAgent status
func (m LaunchdManagerModel) refreshStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.launchdManager.GetStatus()
		if err != nil {
			return err
		}
		return status
	}
}

// executeAction returns a command to execute the selected action
func (m LaunchdManagerModel) executeAction() tea.Cmd {
	if m.status == nil {
		return nil
	}

	m.processing = true

	return func() tea.Msg {
		var err error
		var message string

		switch m.selectedAction {
		case 0: // Load
			if m.status.Loaded {
				return ActionResult{
					Error:   fmt.Errorf("agent is already loaded"),
					Message: "",
				}
			}
			err = m.launchdManager.Load()
			if err == nil {
				message = "LaunchAgent loaded successfully"
			}

		case 1: // Unload
			if !m.status.Loaded {
				return ActionResult{
					Error:   fmt.Errorf("agent is not loaded"),
					Message: "",
				}
			}
			err = m.launchdManager.Unload()
			if err == nil {
				message = "LaunchAgent unloaded successfully"
			}

		case 2: // Start
			if !m.status.Loaded {
				return ActionResult{
					Error:   fmt.Errorf("agent must be loaded first"),
					Message: "",
				}
			}
			if m.status.Running {
				return ActionResult{
					Error:   fmt.Errorf("agent is already running"),
					Message: "",
				}
			}
			err = m.launchdManager.Start()
			if err == nil {
				message = "LaunchAgent started successfully"
			}

		case 3: // Stop
			if !m.status.Running {
				return ActionResult{
					Error:   fmt.Errorf("agent is not running"),
					Message: "",
				}
			}
			err = m.launchdManager.Stop()
			if err == nil {
				message = "LaunchAgent stopped successfully"
			}

		case 4: // Remove
			err = m.launchdManager.Remove()
			if err == nil {
				message = "LaunchAgent removed successfully"
			}

		case 5: // Refresh
			// Just refresh status, no action needed
			return ActionResult{Message: "Status refreshed"}
		}

		if err != nil {
			return ActionResult{
				Error:   err,
				Message: "",
			}
		}

		return ActionResult{
			Error:   nil,
			Message: message,
		}
	}
}

// ActionResult represents the result of a LaunchAgent action
type ActionResult struct {
	Error   error
	Message string
}
