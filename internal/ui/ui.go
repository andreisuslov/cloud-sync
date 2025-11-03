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
	"github.com/andreisuslov/cloud-sync/internal/ui/models"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
	"github.com/andreisuslov/cloud-sync/internal/ui/views"
)

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
}

func (i MenuItem) FilterValue() string { return i.Title }

func (i MenuItem) Title() string       { return i.Title }
func (i MenuItem) Description() string { return i.Description }

// NewModel creates a new application model
func NewModel() models.Model {
	// Create menu items
	items := []list.Item{
		MenuItem{
			Title:       "1. Installation & Setup",
			Description: "Install required tools and configure remotes",
		},
		MenuItem{
			Title:       "2. Backup Operations",
			Description: "Run manual sync or trigger automated backups",
		},
		MenuItem{
			Title:       "3. Log Viewer",
			Description: "View backup logs and transfer history",
		},
		MenuItem{
			Title:       "4. LaunchAgent Management",
			Description: "Manage automated backup scheduling",
		},
		MenuItem{
			Title:       "5. Maintenance",
			Description: "Remove lockfiles, reset timestamps, clear logs",
		},
		MenuItem{
			Title:       "6. Help",
			Description: "View keyboard shortcuts and documentation",
		},
		MenuItem{
			Title:       "7. Exit",
			Description: "Exit the application",
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

	return models.Model{
		State:   models.StateMainMenu,
		List:    l,
		Spinner: s,
	}
}

// Init initializes the model
func (m models.Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

// Update handles messages and updates the model
func (m models.Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m.handleKeyPress(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Delegate to active sub-view if present
	if m.ActiveSubView != nil && m.State != models.StateMainMenu {
		var cmd tea.Cmd
		m.ActiveSubView, cmd = m.ActiveSubView.Update(msg)
		return m, cmd
	}

	// Update the appropriate sub-component based on state
	switch m.State {
	case models.StateMainMenu:
		var cmd tea.Cmd
		m.List, cmd = m.List.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m models.Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Let sub-views handle their own key presses first if they have special handling
	if m.ActiveSubView != nil && m.State != models.StateMainMenu {
		// Only intercept 'q' to go back, let sub-views handle everything else
		if msg.String() == "q" || msg.String() == "esc" {
			// Go back to main menu
			m.State = models.StateMainMenu
			m.ActiveSubView = nil
			return m, nil
		}
		// Let the sub-view handle the key
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		if m.State == models.StateMainMenu {
			m.Quitting = true
			return m, tea.Quit
		}
		// Go back to main menu from other states
		m.State = models.StateMainMenu
		m.ActiveSubView = nil
		return m, nil

	case "enter":
		if m.State == models.StateMainMenu {
			return m.handleMenuSelection()
		}
	}

	return m, nil
}

// handleMenuSelection handles menu item selection
func (m models.Model) handleMenuSelection() (tea.Model, tea.Cmd) {
	selected := m.List.SelectedItem()
	if selected == nil {
		return m, nil
	}

	menuItem := selected.(MenuItem)

	switch {
	case strings.HasPrefix(menuItem.Title, "1."):
		m.State = models.StateInstallation
		m.Message = "Installation wizard not yet implemented"
		m.ShowMessage = true
	case strings.HasPrefix(menuItem.Title, "2."):
		m.State = models.StateBackupRunning
		// Initialize backup operations view
		backupModel := views.NewBackupOpsModel(views.BackupManual, m.Width, m.Height)
		m.ActiveSubView = backupModel
		return m, backupModel.Init()
	case strings.HasPrefix(menuItem.Title, "3."):
		m.State = models.StateLogViewer
		// Initialize log viewer
		homeDir, _ := os.UserHomeDir()
		logDir := filepath.Join(homeDir, "Documents", "rclone_logs")
		logManager := logs.NewManager(logDir)
		logModel := views.NewLogViewerModel(logManager, views.LogModeAll, m.Width, m.Height)
		m.ActiveSubView = logModel
		return m, logModel.Init()
	case strings.HasPrefix(menuItem.Title, "4."):
		m.State = models.StateLaunchdManager
		// Initialize LaunchAgent manager
		username := os.Getenv("USER")
		launchdManager := launchd.NewManager(username)
		launchdModel := views.NewLaunchdManagerModel(launchdManager, m.Width, m.Height)
		m.ActiveSubView = launchdModel
		return m, launchdModel.Init()
	case strings.HasPrefix(menuItem.Title, "5."):
		m.State = models.StateMaintenance
		m.Message = "Maintenance tools not yet implemented"
		m.ShowMessage = true
	case strings.HasPrefix(menuItem.Title, "6."):
		m.State = models.StateHelp
	case strings.HasPrefix(menuItem.Title, "7."):
		m.Quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI
func (m models.Model) View() string {
	if m.Quitting {
		return styles.RenderInfo("\nThank you for using Cloud Sync!\n")
	}

	// If we have an active sub-view, render it
	if m.ActiveSubView != nil && m.State != models.StateMainMenu {
		return m.ActiveSubView.View()
	}

	var content string

	switch m.State {
	case models.StateMainMenu:
		content = m.viewMainMenu()
	case models.StateInstallation:
		content = m.viewPlaceholder("Installation & Setup")
	case models.StateConfiguration:
		content = m.viewPlaceholder("Configuration")
	case models.StateBackupRunning:
		content = m.viewPlaceholder("Backup Operations")
	case models.StateLogViewer:
		content = m.viewPlaceholder("Log Viewer")
	case models.StateLaunchdManager:
		content = m.viewPlaceholder("LaunchAgent Manager")
	case models.StateMaintenance:
		content = m.viewPlaceholder("Maintenance")
	case models.StateHelp:
		content = m.viewHelp()
	default:
		content = "Unknown state"
	}

	return content
}

// viewMainMenu renders the main menu
func (m models.Model) viewMainMenu() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.List.View())
	b.WriteString("\n\n")
	b.WriteString(styles.RenderHelp("↑/↓: Navigate • Enter: Select • q: Quit"))

	return b.String()
}

// viewPlaceholder renders a placeholder view for unimplemented features
func (m models.Model) viewPlaceholder(title string) string {
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
func (m models.Model) viewHelp() string {
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