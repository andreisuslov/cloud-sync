package views

import (
	"fmt"
	"strings"

	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			title:       "3. Install/Update rsync (Optional)",
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
					m.statusMsg = msg.message
				} else {
					m.items[i].status = StatusFailed
					m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
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

	style := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("\n↑/↓ or j/k: navigate (wrap-around) • enter: execute step • q: quit")

	statusText := ""
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Padding(1, 0)
		statusText = statusStyle.Render("\n" + m.statusMsg)
	}

	return style.Render(m.list.View() + helpText + statusText)
}

// executeStep executes a specific installation step
func (m InstallationModel) executeStep(item InstallationItem) tea.Cmd {
	return func() tea.Msg {
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
				err:     fmt.Errorf("step not implemented yet"),
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
		// rsync is installed, try to update
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
			message: "✓ rsync updated successfully",
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
		message: "✓ rsync installed successfully",
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

// GetStatusColor returns a color for the given status
func GetStatusColor(status InstallStatus) lipgloss.Color {
	switch status {
	case StatusPending:
		return lipgloss.Color("241")
	case StatusInProgress:
		return lipgloss.Color("214")
	case StatusComplete:
		return lipgloss.Color("42")
	case StatusFailed:
		return lipgloss.Color("196")
	case StatusSkipped:
		return lipgloss.Color("243")
	default:
		return lipgloss.Color("255")
	}
}

// FormatStepWithStatus formats a step title with its status
func FormatStepWithStatus(title string, status InstallStatus) string {
	icon := GetStatusIcon(status)
	color := GetStatusColor(status)
	style := lipgloss.NewStyle().Foreground(color)
	return fmt.Sprintf("%s %s", style.Render(icon), title)
}
