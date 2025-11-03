package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
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

// keyMap defines key bindings for the application
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
}

// defaultKeyMap returns the default key bindings
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

// ShortHelp returns key bindings for the short help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Quit}
}

// FullHelp returns key bindings for the full help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Back, k.Quit, k.Help},
	}
}

// Model represents the main application model
type Model struct {
	State         AppState
	List          list.Model
	Spinner       spinner.Model
	HelpViewport  viewport.Model
	Help          help.Model
	Keys          keyMap
	Width         int
	Height        int
	Err           error
	Message       string
	ShowMessage   bool
	Quitting      bool
	HelpReady     bool
	
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

	h := help.New()
	h.ShowAll = false

	return Model{
		State:   StateMainMenu,
		List:    l,
		Spinner: s,
		Help:    h,
		Keys:    defaultKeyMap(),
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
		// Give the list most of the vertical space, leaving room for title and help
		listHeight := msg.Height - 8
		if listHeight < 5 {
			listHeight = 5 // Minimum height
		}
		m.List.SetSize(msg.Width-4, listHeight)
		
		// Update help viewport if in help state
		if m.State == StateHelp && m.HelpReady {
			m.HelpViewport.Width = msg.Width - 4
			m.HelpViewport.Height = msg.Height - 6
		}
		
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
			// Check for quit or enter in main menu using key bindings
			if key.Matches(msg, m.Keys.Quit) {
				m.Quitting = true
				return m, tea.Quit
			}
			if key.Matches(msg, m.Keys.Back) && msg.String() == "q" {
				m.Quitting = true
				return m, tea.Quit
			}
			if key.Matches(msg, m.Keys.Enter) {
				return m.handleMenuSelection()
			}
			// Let the list handle navigation keys (up, down, j, k, etc.)
			var cmd tea.Cmd
			m.List, cmd = m.List.Update(msg)
			return m, cmd
		} else if m.State == StateHelp {
			// Handle help screen navigation
			switch msg.String() {
			case "q", "esc":
				m.State = StateMainMenu
				m.HelpReady = false
				return m, nil
			case "ctrl+c":
				m.Quitting = true
				return m, tea.Quit
			default:
				// Let viewport handle scrolling keys
				var cmd tea.Cmd
				m.HelpViewport, cmd = m.HelpViewport.Update(msg)
				return m, cmd
			}
		} else {
			// Handle keys for other states
			return m.handleKeyPress(msg)
		}

	case tea.MouseMsg:
		// Pass mouse events to the list for scrolling in main menu
		if m.State == StateMainMenu {
			var cmd tea.Cmd
			m.List, cmd = m.List.Update(msg)
			return m, cmd
		}
		// Pass mouse events to help viewport
		if m.State == StateHelp {
			var cmd tea.Cmd
			m.HelpViewport, cmd = m.HelpViewport.Update(msg)
			return m, cmd
		}
		// Pass mouse events to active sub-view
		if m.ActiveSubView != nil {
			var cmd tea.Cmd
			m.ActiveSubView, cmd = m.ActiveSubView.Update(msg)
			return m, cmd
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
		// Initialize help viewport with content
		m.HelpViewport = viewport.New(m.Width-4, m.Height-6)
		m.HelpViewport.SetContent(m.getHelpContent())
		m.HelpReady = true
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
	
	// Render help using the help component
	helpView := m.Help.View(m.Keys)
	b.WriteString(helpView)

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

// getHelpContent returns the help text content
func (m Model) getHelpContent() string {
	return `Keyboard Shortcuts & Help
========================

Global Shortcuts:
  q, esc       - Return to main menu / quit
  ctrl+c       - Force quit
  ↑/↓, j/k     - Navigate lists / scroll content
  enter        - Select / confirm
  pgup/pgdn    - Page up / page down
  home/end     - Jump to start / end

Main Menu:
  1-7          - Quick access to menu items

Log Viewer:
  /            - Search
  n            - Next search result
  N            - Previous search result
  g            - Go to top
  G            - Go to bottom
  tab          - Switch between viewing modes
  e            - Export logs

Backup Operations:
  ctrl+c       - Cancel running backup
  r            - Retry failed backup

LaunchAgent Manager:
  l            - Load agent
  u            - Unload agent
  s            - Start agent
  t            - Stop agent
  r            - Remove agent

Installation & Configuration:
  Follow on-screen prompts
  Tab to navigate between fields
  Enter to confirm selections

Tips:
  - Use arrow keys or j/k to scroll in any view
  - Mouse scrolling is supported throughout
  - Press 'q' or 'esc' to return to previous screen
  - All changes are saved automatically

Project Information:
  GitHub: https://github.com/andreisuslov/cloud-sync
  Author: andreisuslov
  
For more information and documentation, visit the GitHub repository.`
}

// viewHelp renders the help screen with scrollable viewport
func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.RenderTitle("Help & Documentation"))
	b.WriteString("\n\n")
	
	if m.HelpReady {
		b.WriteString(m.HelpViewport.View())
		b.WriteString("\n\n")
		
		// Show scroll position indicator
		scrollInfo := fmt.Sprintf("%.0f%%", m.HelpViewport.ScrollPercent()*100)
		b.WriteString(styles.RenderMuted(fmt.Sprintf("Scroll: %s | ", scrollInfo)))
	}
	
	b.WriteString(styles.RenderHelp("↑/↓, j/k: Scroll • q: Back to Main Menu"))

	return b.String()
}