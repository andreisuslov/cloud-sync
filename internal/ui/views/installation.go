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

	// Output box styles
	outputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			MarginTop(1)

	outputTitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	outputContentStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)
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

// ConfigurationSetupModel represents the configuration setup view state
type ConfigurationSetupModel struct {
	list         list.Model
	items        []InstallationItem
	currentStep  int
	width        int
	height       int
	err          error
	quitting     bool
	installer    *installer.Installer
	statusMsg    string
	outputBuffer []string // Buffer to store command output
	showOutput   bool     // Whether to show the output box
}

// NewConfigurationSetupModel creates a new configuration setup model
func NewConfigurationSetupModel() ConfigurationSetupModel {
	items := []InstallationItem{
		{
			title:       "1. Check rclone Installation",
			description: "Verify that rclone is available on your system",
			status:      StatusPending,
		},
		{
			title:       "2. Check rclone Version",
			description: "Ensure rclone version is compatible (≥1.50)",
			status:      StatusPending,
		},
		{
			title:       "3. Install/Update rclone",
			description: "Install latest rclone version via Homebrew if needed",
			status:      StatusPending,
		},
		{
			title:       "4. Manage Remotes",
			description: "Setup cloud storage remotes (rclone config)",
			status:      StatusPending,
		},
		{
			title:       "5. List Configured Remotes",
			description: "View all configured cloud storage remotes",
			status:      StatusPending,
		},
		{
			title:       "6. Test Remote Connection",
			description: "Verify connectivity to configured remotes",
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
	l.Title = "Configuration (rclone config)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return ConfigurationSetupModel{
		list:         l,
		items:        items,
		currentStep:  0,
		width:        80,
		height:       20,
		installer:    installer.NewInstaller(),
		outputBuffer: make([]string, 0),
		showOutput:   false,
	}
}

// Init initializes the configuration setup model
func (m ConfigurationSetupModel) Init() tea.Cmd {
	// Request window size on initialization
	return tea.ClearScreen
}

// Update handles messages for the configuration setup view
func (m ConfigurationSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust list size based on whether output is shown
		listHeight := msg.Height - 4
		if m.showOutput {
			// Split screen: 60% for list, 40% for output
			listHeight = int(float64(msg.Height) * 0.5)
		}
		m.list.SetSize(msg.Width-4, listHeight)
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
				
				// Add output to buffer if present
				if msg.output != "" {
					m.showOutput = true
					m.outputBuffer = append(m.outputBuffer, msg.output)
					// Keep only last 50 lines
					if len(m.outputBuffer) > 50 {
						m.outputBuffer = m.outputBuffer[len(m.outputBuffer)-50:]
					}
				}
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

// View renders the configuration setup view
func (m ConfigurationSetupModel) View() string {
	if m.quitting {
		return "Configuration cancelled.\n"
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

	mainContent := baseStyle.Render(m.list.View() + helpText + statusText)

	// Add output box if there's output to show
	if m.showOutput && len(m.outputBuffer) > 0 {
		outputTitle := outputTitleStyle.Render("Command Output:")
		
		// Join output lines and limit display
		outputLines := m.outputBuffer
		if len(outputLines) > 15 {
			outputLines = outputLines[len(outputLines)-15:]
		}
		outputText := outputContentStyle.Render(strings.Join(outputLines, "\n"))
		
		outputBox := outputBoxStyle.Width(m.width - 8).Render(outputTitle + "\n" + outputText)
		return lipgloss.JoinVertical(lipgloss.Left, mainContent, outputBox)
	}

	return mainContent
}

// executeStep executes a specific configuration step
func (m ConfigurationSetupModel) executeStep(item InstallationItem) tea.Cmd {
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
		case strings.Contains(item.title, "Check rclone Installation"):
			return m.checkRcloneInstallation(item)
		case strings.Contains(item.title, "Check rclone Version"):
			return m.checkRcloneVersion(item)
		case strings.Contains(item.title, "Install/Update rclone"):
			return m.installOrUpdateRclone(item)
		case strings.Contains(item.title, "Manage Remotes"):
			return m.manageRemotes(item)
		case strings.Contains(item.title, "List Configured Remotes"):
			return m.listRemotes(item)
		case strings.Contains(item.title, "Test Remote Connection"):
			return m.testRemoteConnection(item)
		default:
			return installStepCompleteMsg{
				step:    item.title,
				success: false,
				message: "✗ Step not implemented yet",
			}
		}
	}
}

// checkRcloneInstallation checks if rclone is installed
func (m ConfigurationSetupModel) checkRcloneInstallation(item InstallationItem) installStepCompleteMsg {
	if m.installer.CheckRcloneInstalled() {
		path, err := m.installer.GetRclonePath()
		if err != nil {
			return installStepCompleteMsg{
				step:    item.title,
				success: true,
				message: "✓ rclone is installed",
			}
		}
		output := fmt.Sprintf("rclone binary location:\n%s", path)
		return installStepCompleteMsg{
			step:    item.title,
			success: true,
			message: fmt.Sprintf("✓ rclone is installed"),
			output:  output,
		}
	}
	return installStepCompleteMsg{
		step:    item.title,
		success: false,
		err:     fmt.Errorf("rclone is not installed"),
		message: "✗ rclone is not installed",
	}
}

// checkRcloneVersion checks the rclone version
func (m ConfigurationSetupModel) checkRcloneVersion(item InstallationItem) installStepCompleteMsg {
	version, output, err := m.installer.GetRcloneVersionWithOutput()
	if err != nil {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     err,
			message: fmt.Sprintf("✗ Failed to get rclone version: %v", err),
			output:  output,
		}
	}
	return installStepCompleteMsg{
		step:    item.title,
		success: true,
		message: fmt.Sprintf("✓ %s", version),
		output:  output,
	}
}

// installOrUpdateRclone installs or updates rclone
func (m ConfigurationSetupModel) installOrUpdateRclone(item InstallationItem) installStepCompleteMsg {
	if !m.installer.CheckHomebrewInstalled() {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     fmt.Errorf("homebrew is required but not installed"),
			message: "✗ Homebrew is required. Please install Homebrew first.",
		}
	}

	if m.installer.CheckRcloneInstalled() {
		// rclone is installed, try to update it
		output, err := m.installer.UpdateRcloneWithOutput()
		if err != nil {
			// Update might fail if already up-to-date
			if strings.Contains(err.Error(), "already installed") {
				return installStepCompleteMsg{
					step:    item.title,
					success: true,
					message: "✓ rclone is already up-to-date",
					output:  output,
				}
			}
			return installStepCompleteMsg{
				step:    item.title,
				success: false,
				err:     err,
				message: fmt.Sprintf("✗ Failed to update rclone: %v", err),
				output:  output,
			}
		}
		return installStepCompleteMsg{
			step:    item.title,
			success: true,
			message: "✓ rclone updated successfully via Homebrew",
			output:  output,
		}
	}

	// rclone is not installed, install it
	output, err := m.installer.InstallRcloneWithOutput()
	if err != nil {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     err,
			message: fmt.Sprintf("✗ Failed to install rclone: %v", err),
			output:  output,
		}
	}
	return installStepCompleteMsg{
		step:    item.title,
		success: true,
		message: "✓ rclone installed successfully via Homebrew",
		output:  output,
	}
}

// manageRemotes opens the rclone config interactive wizard
func (m ConfigurationSetupModel) manageRemotes(item InstallationItem) installStepCompleteMsg {
	if !m.installer.CheckRcloneInstalled() {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     fmt.Errorf("rclone is not installed"),
			message: "✗ rclone must be installed first",
		}
	}

	output, err := m.installer.RunRcloneConfig()
	if err != nil {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     err,
			message: fmt.Sprintf("✗ Failed to run rclone config: %v", err),
			output:  output,
		}
	}

	return installStepCompleteMsg{
		step:    item.title,
		success: true,
		message: "✓ rclone config completed",
		output:  output,
	}
}

// listRemotes lists all configured rclone remotes
func (m ConfigurationSetupModel) listRemotes(item InstallationItem) installStepCompleteMsg {
	if !m.installer.CheckRcloneInstalled() {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     fmt.Errorf("rclone is not installed"),
			message: "✗ rclone must be installed first",
		}
	}

	remotes, err := m.installer.ListRcloneRemotes()
	if err != nil {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     err,
			message: fmt.Sprintf("✗ Failed to list remotes: %v", err),
		}
	}

	if len(remotes) == 0 {
		return installStepCompleteMsg{
			step:    item.title,
			success: true,
			message: "✓ No remotes configured yet. Use 'Manage Remotes' to add one.",
		}
	}

	output := fmt.Sprintf("Configured remotes:\n%s", strings.Join(remotes, "\n"))
	return installStepCompleteMsg{
		step:    item.title,
		success: true,
		message: fmt.Sprintf("✓ Found %d configured remote(s)", len(remotes)),
		output:  output,
	}
}

// testRemoteConnection tests connectivity to a configured remote
func (m ConfigurationSetupModel) testRemoteConnection(item InstallationItem) installStepCompleteMsg {
	if !m.installer.CheckRcloneInstalled() {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     fmt.Errorf("rclone is not installed"),
			message: "✗ rclone must be installed first",
		}
	}

	remotes, err := m.installer.ListRcloneRemotes()
	if err != nil || len(remotes) == 0 {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			message: "✗ No remotes configured. Use 'Manage Remotes' first.",
		}
	}

	// Test the first remote
	firstRemote := strings.TrimSuffix(remotes[0], ":")
	output, err := m.installer.TestRcloneRemote(firstRemote)
	if err != nil {
		return installStepCompleteMsg{
			step:    item.title,
			success: false,
			err:     err,
			message: fmt.Sprintf("✗ Failed to connect to remote '%s': %v", firstRemote, err),
			output:  output,
		}
	}

	return installStepCompleteMsg{
		step:    item.title,
		success: true,
		message: fmt.Sprintf("✓ Successfully connected to remote '%s'", firstRemote),
		output:  output,
	}
}

// installStepCompleteMsg is sent when a configuration step completes
type installStepCompleteMsg struct {
	step    string
	success bool
	err     error
	message string
	output  string // Command output to display in the output box
}

// UpdateStepStatus updates the status of a specific step
func (m *ConfigurationSetupModel) UpdateStepStatus(stepIndex int, status InstallStatus) {
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
