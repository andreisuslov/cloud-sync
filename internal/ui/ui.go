package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/launchd"
	"github.com/andreisuslov/cloud-sync/internal/logs"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
	"github.com/andreisuslov/cloud-sync/internal/ui/views"
)

// AppState represents the current state of the application
type AppState int

const (
	StateMainMenu AppState = iota
	StateInstallation
	StateConfiguration
	StateBackupRunning
	StateLogViewer
	StateLaunchdManager
	StateMaintenance
	StateHelp
	StateExiting
)

// Model represents the main application model
type Model struct {
	State         AppState
	List          list.Model
	Spinner       spinner.Model
	Width         int
	Height        int
	Err           error
	Message       string
	ShowMessage   bool
	Quitting      bool
	
	// Active sub-view (when navigated to a specific view)
	ActiveSubView tea.Model
}

// MenuItem represents a menu item
type MenuItem struct {
	title       string
	description string
}

func (i MenuItem) FilterValue() string { return i.title }

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Description() string { return i.description }

// NewModel creates a new application model
func NewModel() Model {
	// Create menu items
	items := []list.Item{
		MenuItem{
			title:       "1. Installation & Setup",
			description: "Install required tools and configure remotes",
		},
		MenuItem{
			title:       "2. Backup Operations",
			description: "Run manual sync or trigger automated backups",
		},
		MenuItem{
			title:       "3. Log Viewer",
			description: "View backup logs and transfer history",
		},
		MenuItem{
			title:       "4. LaunchAgent Management",
			description: "Manage automated backup scheduling",
		},
		MenuItem{
			title:       "5. Maintenance",
			description: "Remove lockfiles, reset timestamps, clear logs",
		},
		MenuItem{
			title:       "6. Help",
			description: "View keyboard shortcuts and documentation",
		},
		MenuItem{
			title:       "7. Exit",
			description: "Exit the application",
		},
	}

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Cloud Sync - Backup Management"
	l.Styles.Title = styles.TitleStyle

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	return Model{
		State:   StateMainMenu,
		List:    l,
		Spinner: s,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.List.SetSize(msg.Width-4, msg.Height-8)
		// Also update active sub-view if present
		if m.ActiveSubView != nil {
			var cmd tea.Cmd
			m.ActiveSubView, cmd = m.ActiveSubView.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// Handle special keys that should be intercepted
		if m.State == StateMainMenu {
			// Check for quit or enter in main menu
			switch msg.String() {
			case "ctrl+c", "q":
				m.Quitting = true
				return m, tea.Quit
			case "enter":
				return m.handleMenuSelection()
			}
			// Let the list handle navigation keys (up, down, j, k, etc.)
			var cmd tea.Cmd
			m.List, cmd = m.List.Update(msg)
			return m, cmd
		} else {
			// Handle keys for other states
			return m.handleKeyPress(msg)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Delegate to active sub-view if present
	if m.ActiveSubView != nil && m.State != StateMainMenu {
		var cmd tea.Cmd
		m.ActiveSubView, cmd = m.ActiveSubView.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKeyPress handles keyboard input for non-main-menu states
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle 'q' or 'esc' to go back to main menu
	if msg.String() == "q" || msg.String() == "esc" {
		m.State = StateMainMenu
		m.ActiveSubView = nil
		return m, nil
	}
	
	// Handle ctrl+c to force quit
	if msg.String() == "ctrl+c" {
		m.Quitting = true
		return m, tea.Quit
	}

	// If we have an active sub-view, let it handle the key
	if m.ActiveSubView != nil {
		var cmd tea.Cmd
		m.ActiveSubView, cmd = m.ActiveSubView.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleMenuSelection handles menu item selection
func (m Model) handleMenuSelection() (tea.Model, tea.Cmd) {
	selected := m.List.SelectedItem()
	if selected == nil {
		return m, nil
	}

	menuItem := selected.(MenuItem)
	title := menuItem.Title()

	switch {
	case strings.HasPrefix(title, "1."):
		m.State = StateInstallation
		m.Message = "Installation wizard not yet implemented"
		m.ShowMessage = true
	case strings.HasPrefix(title, "2."):
		m.State = StateBackupRunning
		// Initialize backup operations view
		backupModel := views.NewBackupOpsModel(views.BackupManual, m.Width, m.Height)
		m.ActiveSubView = backupModel
		return m, backupModel.Init()
	case strings.HasPrefix(title, "3."):
		m.State = StateLogViewer
		// Initialize log viewer
		homeDir, _ := os.UserHomeDir()
		logDir := filepath.Join(homeDir, "Documents", "rclone_logs")
		logManager := logs.NewManager(logDir)
		logModel := views.NewLogViewerModel(logManager, views.LogViewAll, m.Width, m.Height)
		m.ActiveSubView = logModel
		return m, logModel.Init()
	case strings.HasPrefix(title, "4."):
		m.State = StateLaunchdManager
		// Initialize LaunchAgent manager
		username := os.Getenv("USER")
		launchdManager := launchd.NewManager(username)
		launchdModel := views.NewLaunchdManagerModel(launchdManager, m.Width, m.Height)
		m.ActiveSubView = launchdModel
		return m, launchdModel.Init()
	case strings.HasPrefix(title, "5."):
		m.State = StateMaintenance
		m.Message = "Maintenance tools not yet implemented"
		m.ShowMessage = true
	case strings.HasPrefix(title, "6."):
		m.State = StateHelp
	case strings.HasPrefix(title, "7."):
		m.Quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.Quitting {
		return styles.RenderInfo("\nThank you for using Cloud Sync!\n")
	}

	// If we have an active sub-view, render it
	if m.ActiveSubView != nil && m.State != StateMainMenu {
		return m.ActiveSubView.View()
	}

	var content string

	switch m.State {
	case StateMainMenu:
		content = m.viewMainMenu()
	case StateInstallation:
		content = m.viewPlaceholder("Installation & Setup")
	case StateConfiguration:
		content = m.viewPlaceholder("Configuration")
	case StateBackupRunning:
		content = m.viewPlaceholder("Backup Operations")
	case StateLogViewer:
		content = m.viewPlaceholder("Log Viewer")
	case StateLaunchdManager:
		content = m.viewPlaceholder("LaunchAgent Manager")
	case StateMaintenance:
		content = m.viewPlaceholder("Maintenance")
	case StateHelp:
		content = m.viewHelp()
	default:
		content = "Unknown state"
	}

	return content
}

// viewMainMenu renders the main menu
func (m Model) viewMainMenu() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.List.View())
	b.WriteString("\n\n")
	b.WriteString(styles.RenderHelp("↑/↓: Navigate • Enter: Select • q: Quit"))

	return b.String()
}

// viewPlaceholder renders a placeholder view for unimplemented features
func (m Model) viewPlaceholder(title string) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.RenderTitle(title))
	b.WriteString("\n\n")

	if m.ShowMessage {
		b.WriteString(styles.RenderWarning(m.Message))
		b.WriteString("\n\n")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(2, 4).
		Width(60).
		Render(fmt.Sprintf(
			"%s\n\nThis feature is under development.\n\nPress 'q' to return to main menu.",
			m.Spinner.View(),
		))

	b.WriteString(box)
	b.WriteString("\n\n")
	b.WriteString(styles.RenderHelp("q: Back to Main Menu • ctrl+c: Quit"))

	return b.String()
}

// viewHelp renders the help screen
func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.RenderTitle("Keyboard Shortcuts & Help"))
	b.WriteString("\n\n")

	help := `Global Shortcuts:
  q, esc       - Return to main menu / quit
  ctrl+c       - Force quit
  ↑/↓, j/k     - Navigate lists
  enter        - Select / confirm

Main Menu:
  1-7          - Quick access to menu items

Log Viewer:
  /            - Search
  n            - Next search result
  N            - Previous search result
  g            - Go to top
  G            - Go to bottom

Backup Operations:
  ctrl+c       - Cancel running backup
  r            - Retry failed backup

For more information, visit:
https://github.com/andreisuslov/cloud-sync`

	box := styles.RenderBox(help)
	b.WriteString(box)
	b.WriteString("\n\n")
	b.WriteString(styles.RenderHelp("q: Back to Main Menu"))

	return b.String()
}