package views

import (
	"fmt"
	"strings"

	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style definitions using lipgloss
var (
	// Color palette
	colorPrimary   = lipgloss.Color("62")
	colorSecondary = lipgloss.Color("241")
	colorSuccess   = lipgloss.Color("42")
	colorWarning   = lipgloss.Color("214")
	colorError     = lipgloss.Color("196")
	colorMuted     = lipgloss.Color("243")
	colorWhite     = lipgloss.Color("255")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Padding(1, 0).
			MarginTop(1)

	successStatusStyle = statusStyle.Copy().
				Foreground(colorSuccess)

	warningStatusStyle = statusStyle.Copy().
				Foreground(colorWarning)

	errorStatusStyle = statusStyle.Copy().
				Foreground(colorError)

	// Status icon styles
	pendingStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	inProgressStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	completeStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	failedStyle = lipgloss.NewStyle().
			Foreground(colorError)

	skippedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// InstallationItem represents an installation step
type InstallationItem struct {
	title       string
	description string
	status      InstallStatus
}

// InstallStatus represents the status of an installation step
type InstallStatus int

const (
	StatusPending InstallStatus = iota
	StatusInProgress
	StatusComplete
	StatusFailed
	StatusSkipped
)

func (i InstallationItem) Title() string       { return i.title }
func (i InstallationItem) Description() string { return i.description }
func (i InstallationItem) FilterValue() string { return i.title }

// InstallationModel represents the installation view state
type InstallationModel struct {
	list         list.Model
	items        []InstallationItem
	currentStep  int
	width        int
	height       int
	err          error
	quitting     bool
	installer    *installer.Installer
	statusMsg    string
}

// NewInstallationModel creates a new installation model
func NewInstallationModel() InstallationModel {
	items := []InstallationItem{
		{
			title:       "1. Check rsync Installation",
			description: "Verify that rsync is available on your system",
			status:      StatusPending,
		},
		{
			title:       "2. Check rsync Version",
			description: "Ensure rsync version is compatible (≥3.0)",
			status:      StatusPending,
		},
		{
			title:       "3. Install/Update rsync",
			description: "Install latest rsync version via Homebrew if needed",
			status:      StatusPending,
		},
		{
			title:       "4. Verify SSH Access",
			description: "Test SSH connectivity to remote hosts",
			status:      StatusPending,
		},
		{
			title:       "5. Setup SSH Keys",
			description: "Configure passwordless SSH authentication",
			status:      StatusPending,
		},
		{
			title:       "6. Configure rsync Profiles",
			description: "Setup source and destination paths",
			status:      StatusPending,
		},
		{
			title:       "7. Test rsync Connection",
			description: "Verify rsync can sync with remote host",
			status:      StatusPending,
		},
	}

	// Convert to list items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	// Set reasonable default dimensions (will be updated on first WindowSizeMsg)
	l := list.New(listItems, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Installation & Setup"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return InstallationModel{
		list:        l,
		items:       items,
		currentStep: 0,
		width:       80,
		height:      20,
		installer:   installer.NewInstaller(),
	}
}

// Init initializes the installation model
func (m InstallationModel) Init() tea.Cmd {
	// Request window size on initialization
	return tea.ClearScreen
}

// Update handles messages for the installation view
func (m InstallationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-4)
		return m, nil

	case installStepCompleteMsg:
		// Update the step status based on the result
		for i := range m.items {
			if m.items[i].title == msg.step {
				if msg.success {
					m.items[i].status = StatusComplete
					if msg.message != "" {
						m.statusMsg = msg.message
					}
				} else {
					m.items[i].status = StatusFailed
					// Use message if available, otherwise format error
					if msg.message != "" {
						m.statusMsg = msg.message
					} else if msg.err != nil {
						m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
					} else {
						m.statusMsg = "Error: Unknown error occurred"
					}
				}
				m.UpdateStepStatus(i, m.items[i].status)
				break
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// Execute the selected installation step
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				item := selectedItem.(InstallationItem)
				return m, m.executeStep(item)
			}

		case "up", "k":
			// Wrap around to bottom when at top
			if m.list.Index() == 0 {
				m.list.Select(len(m.items) - 1)
				return m, nil
			}

		case "down", "j":
			// Wrap around to top when at bottom
			if m.list.Index() == len(m.items)-1 {
				m.list.Select(0)
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the installation view
func (m InstallationModel) View() string {
	if m.quitting {
		return "Installation cancelled.\n"
	}

	helpText := helpStyle.Render("\n↑/↓ or j/k: navigate (wrap-around) • enter: execute step • q: quit")

	statusText := ""
	if m.statusMsg != "" {
		// Determine which status style to use based on message content
		var msgStyle lipgloss.Style
		if strings.HasPrefix(m.statusMsg, "✓") {
			msgStyle = successStatusStyle
		} else if strings.HasPrefix(m.statusMsg, "✗") {
			msgStyle = errorStatusStyle
		} else {
			msgStyle = warningStatusStyle
		}
		statusText = msgStyle.Render("\n" + m.statusMsg)
	}

	return baseStyle.Render(m.list.View() + helpText + statusText)
}

// executeStep executes a specific installation step
func (m InstallationModel) executeStep(item InstallationItem) tea.Cmd {
	return func() tea.Msg {
		// Add panic recovery to prevent UI crashes
		defer func() {
			if r := recover(); r != nil {
				// If panic occurs, return error message
				_ = installStepCompleteMsg{
					step:    item.title,
					success: false,
					message: fmt.Sprintf("✗ Panic occurred: %v", r),
				}
			}
		}()

		// Determine which step to execute based on title
		switch {
		case strings.Contains(item.title, "Check rsync Installation"):
			return m.checkRsyncInstallation(item)
		case strings.Contains(item.title, "Check rsync Version"):
			return m.checkRsyncVersion(item)
		case strings.Contains(item.title, "Install/Update rsync"):
			return m.installOrUpdateRsync(item)
		default:
			return installStepCompleteMsg{
				step:    item.title,
				success: false,
				message: "✗ Step not implemented yet",
			}
		}
	}
}

// checkRsyncInstallation checks if rsync is installed
func (m InstallationModel) checkRsyncInstallation(item InstallationItem) installStepCompleteMsg {
	if m.installer.CheckRsyncInstalled() {
		path, err := m.installer.GetRsyncPath()
		if err != nil {
			return installStepCompleteMsg{
				step:    item.title,
				success: true,
				message: "✓ rsync is installed",
			}
		}
		return installStepCompleteMsg{
			step:    item.title,
			success: true,
			message: fmt.Sprintf("✓ rsync is installed at: %s", path),
		}
	}
	return installStepCompleteMsg{
		step:    item.title,
		success: false,
		err:     fmt.Errorf("rsync is not installed"),
		message: "✗ rsync is not installed",
	}
}

// checkRsyncVersion checks the rsync version
func (m InstallationModel) checkRsyncVersion(item InstallationItem) installStepCompleteMsg {
	version, err := m.installer.GetRsyncVersion()
	if err != nil {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     err,
			message: fmt.Sprintf("✗ Failed to get rsync version: %v", err),
		}
	}
	return installStepCompleteMsg{
		step:    item.title,
		success: true,
		message: fmt.Sprintf("✓ %s", version),
	}
}

// installOrUpdateRsync installs or updates rsync
func (m InstallationModel) installOrUpdateRsync(item InstallationItem) installStepCompleteMsg {
	if !m.installer.CheckHomebrewInstalled() {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     fmt.Errorf("homebrew is required but not installed"),
			message: "✗ Homebrew is required. Please install Homebrew first.",
		}
	}

	if m.installer.CheckRsyncInstalled() {
		// rsync is installed, check if it's via Homebrew
		if m.installer.IsRsyncInstalledViaHomebrew() {
			// Try to update Homebrew version
			err := m.installer.UpdateRsync()
			if err != nil {
				return installStepCompleteMsg{
					step:    item.title,
					success: false,
					err:     err,
					message: fmt.Sprintf("✗ Failed to update rsync: %v", err),
				}
			}
			return installStepCompleteMsg{
				step:    item.title,
				success: true,
				message: "✓ rsync updated successfully via Homebrew",
			}
		}

		// System rsync detected, offer to install Homebrew version
		err := m.installer.InstallRsync()
		if err != nil {
			return installStepCompleteMsg{
				step:    item.title,
				success: false,
				err:     err,
				message: fmt.Sprintf("✗ System rsync detected. Failed to install Homebrew version: %v", err),
			}
		}
		return installStepCompleteMsg{
			step:    item.title,
			success: true,
			message: "✓ Homebrew rsync installed (system version still available as fallback)",
		}
	}

	// rsync is not installed, install it
	err := m.installer.InstallRsync()
	if err != nil {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     err,
			message: fmt.Sprintf("✗ Failed to install rsync: %v", err),
		}
	}
	return installStepCompleteMsg{
		step:    item.title,
		success: true,
		message: "✓ rsync installed successfully via Homebrew",
	}
}

// installStepCompleteMsg is sent when an installation step completes
type installStepCompleteMsg struct {
	step    string
	success bool
	err     error
	message string
}

// UpdateStepStatus updates the status of a specific step
func (m *InstallationModel) UpdateStepStatus(stepIndex int, status InstallStatus) {
	if stepIndex >= 0 && stepIndex < len(m.items) {
		m.items[stepIndex].status = status
		
		// Update the list item
		listItems := make([]list.Item, len(m.items))
		for i, item := range m.items {
			listItems[i] = item
		}
		m.list.SetItems(listItems)
	}
}

// GetStatusIcon returns an icon for the given status
func GetStatusIcon(status InstallStatus) string {
	switch status {
	case StatusPending:
		return "⏸"
	case StatusInProgress:
		return "⏳"
	case StatusComplete:
		return "✓"
	case StatusFailed:
		return "✗"
	case StatusSkipped:
		return "⊘"
	default:
		return "?"
	}
}

// GetStatusStyle returns a lipgloss style for the given status
func GetStatusStyle(status InstallStatus) lipgloss.Style {
	switch status {
	case StatusPending:
		return pendingStyle
	case StatusInProgress:
		return inProgressStyle
	case StatusComplete:
		return completeStyle
	case StatusFailed:
		return failedStyle
	case StatusSkipped:
		return skippedStyle
	default:
		return lipgloss.NewStyle().Foreground(colorWhite)
	}
}

// FormatStepWithStatus formats a step title with its status
func FormatStepWithStatus(title string, status InstallStatus) string {
	icon := GetStatusIcon(status)
	style := GetStatusStyle(status)
	return fmt.Sprintf("%s %s", style.Render(icon), title)
}
