package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/andreisuslov/cloud-sync/internal/logs"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogViewMode represents different log viewing modes
type LogViewMode int

const (
	LogViewAll LogViewMode = iota
	LogViewToday
	LogViewRecent
	LogViewSessions
	LogViewStats
)

// LogViewerModel represents the log viewer model
type LogViewerModel struct {
	logManager    *logs.Manager
	mode          LogViewMode
	viewport      viewport.Model
	sessionsTable table.Model
	content       string
	width         int
	height        int
	ready         bool
	err           error
}

// NewLogViewerModel creates a new log viewer model
func NewLogViewerModel(logManager *logs.Manager, mode LogViewMode, width, height int) LogViewerModel {
	vp := viewport.New(width-4, height-10)
	vp.Style = styles.ViewportStyle

	// Create sessions table
	columns := []table.Column{
		{Title: "Date/Time", Width: 20},
		{Title: "Type", Width: 12},
		{Title: "Status", Width: 10},
		{Title: "Files", Width: 8},
		{Title: "Duration", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return LogViewerModel{
		logManager:    logManager,
		mode:          mode,
		viewport:      vp,
		sessionsTable: t,
		width:         width,
		height:        height,
		ready:         false,
	}
}

// Init implements tea.Model
func (m LogViewerModel) Init() tea.Cmd {
	return m.loadContent()
}

// Update implements tea.Model
func (m LogViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "1":
			m.mode = LogViewAll
			return m, m.loadContent()
		case "2":
			m.mode = LogViewToday
			return m, m.loadContent()
		case "3":
			m.mode = LogViewRecent
			return m, m.loadContent()
		case "4":
			m.mode = LogViewSessions
			return m, m.loadContent()
		case "5":
			m.mode = LogViewStats
			return m, m.loadContent()
		case "r":
			return m, m.loadContent()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-10)
			m.viewport.Style = styles.ViewportStyle
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 10
		}

	case string:
		// Content loaded
		m.content = msg
		m.viewport.SetContent(m.content)
		m.viewport.GotoTop()
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m LogViewerModel) View() string {
	var b strings.Builder

	helper := NewViewHelper(m.width, m.height)

	// Header
	title := "Log Viewer"
	subtitle := m.getModeDescription()
	b.WriteString(helper.RenderHeader(title, subtitle))

	// Error display
	if m.err != nil {
		b.WriteString(styles.RenderError("Error: " + m.err.Error()))
		b.WriteString("\n")
	}

	// Content viewport
	if m.ready {
		b.WriteString(m.viewport.View())
	} else {
		b.WriteString(styles.RenderInfo("Loading logs..."))
	}

	// Footer
	helpText := "1-5: Switch view • r: Refresh • ↑/↓: Scroll • q/esc: Back"
	b.WriteString(helper.RenderFooter(helpText))

	return b.String()
}

// getModeDescription returns a description of the current view mode
func (m LogViewerModel) getModeDescription() string {
	switch m.mode {
	case LogViewAll:
		return "All transfers"
	case LogViewToday:
		return "Today's transfers"
	case LogViewRecent:
		return "Recent 50 transfers"
	case LogViewSessions:
		return "Sync sessions"
	case LogViewStats:
		return "Statistics"
	default:
		return ""
	}
}

// loadContent returns a command to load log content
func (m LogViewerModel) loadContent() tea.Cmd {
	return func() tea.Msg {
		switch m.mode {
		case LogViewAll:
			return m.renderAllTransfers()
		case LogViewToday:
			return m.renderTodaysTransfers()
		case LogViewRecent:
			return m.renderRecentTransfers()
		case LogViewSessions:
			return m.renderSessions()
		case LogViewStats:
			return m.renderStats()
		default:
			return "Unknown view mode"
		}
	}
}

// renderAllTransfers renders all transfers
func (m LogViewerModel) renderAllTransfers() tea.Msg {
	transfers, err := m.logManager.GetAllTransfers()
	if err != nil {
		return err
	}

	if len(transfers) == 0 {
		return "No transfers found in logs."
	}

	var b strings.Builder
	b.WriteString(styles.RenderInfo(fmt.Sprintf("Total transfers: %d", len(transfers))))
	b.WriteString("\n\n")

	// Group by date
	transfersByDate := make(map[string][]logs.Transfer)
	for _, t := range transfers {
		date := t.Timestamp.Format("2006-01-02")
		transfersByDate[date] = append(transfersByDate[date], t)
	}

	// Render each date group
	for date, dateTransfers := range transfersByDate {
		b.WriteString(styles.RenderSubtitle(fmt.Sprintf("📅 %s (%d files)", date, len(dateTransfers))))
		b.WriteString("\n")

		for _, t := range dateTransfers {
			b.WriteString(fmt.Sprintf("  %s  %s\n", 
				t.Timestamp.Format("15:04:05"),
				t.Filename,
			))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderTodaysTransfers renders today's transfers
func (m LogViewerModel) renderTodaysTransfers() tea.Msg {
	transfers, err := m.logManager.GetTodaysTransfers()
	if err != nil {
		return err
	}

	if len(transfers) == 0 {
		return "No transfers today."
	}

	var b strings.Builder
	b.WriteString(styles.RenderInfo(fmt.Sprintf("Today's transfers: %d", len(transfers))))
	b.WriteString("\n\n")

	for _, t := range transfers {
		b.WriteString(fmt.Sprintf("%s  %s  %s\n", 
			t.Timestamp.Format("15:04:05"),
			styles.RenderSuccess("✓"),
			t.Filename,
		))
	}

	return b.String()
}

// renderRecentTransfers renders recent transfers
func (m LogViewerModel) renderRecentTransfers() tea.Msg {
	transfers, err := m.logManager.GetRecentTransfers(50)
	if err != nil {
		return err
	}

	if len(transfers) == 0 {
		return "No recent transfers found."
	}

	var b strings.Builder
	b.WriteString(styles.RenderInfo(fmt.Sprintf("Recent transfers: %d", len(transfers))))
	b.WriteString("\n\n")

	for i := len(transfers) - 1; i >= 0; i-- {
		t := transfers[i]
		relativeTime := formatRelativeTime(time.Since(t.Timestamp))
		b.WriteString(fmt.Sprintf("%s  %s  %s\n", 
			relativeTime,
			styles.RenderSuccess("✓"),
			t.Filename,
		))
	}

	return b.String()
}

// renderSessions renders sync sessions using table
func (m LogViewerModel) renderSessions() tea.Msg {
	sessions, err := m.logManager.GetSyncSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		return "No sync sessions found."
	}

	// Build table rows
	rows := []table.Row{}
	for i := len(sessions) - 1; i >= 0; i-- {
		session := sessions[i]
		
		dateTime := session.StartTime.Format("2006-01-02 15:04")
		
		status := "✗ Failed"
		if session.Success {
			status = "✓ Success"
		}
		
		duration := "In progress"
		if !session.EndTime.IsZero() {
			d := session.EndTime.Sub(session.StartTime)
			duration = formatDuration(d)
		}
		
		rows = append(rows, table.Row{
			dateTime,
			session.Type,
			status,
			fmt.Sprintf("%d", session.Transfers),
			duration,
		})
	}

	m.sessionsTable.SetRows(rows)

	var b strings.Builder
	b.WriteString(styles.RenderInfo(fmt.Sprintf("Total sessions: %d", len(sessions))))
	b.WriteString("\n\n")
	b.WriteString(m.sessionsTable.View())

	return b.String()
}

// renderSession renders a single sync session
func (m LogViewerModel) renderSession(session logs.SyncSession) string {
	var b strings.Builder

	// Session header
	statusIcon := "✗"
	statusStyle := styles.RenderError
	if session.Success {
		statusIcon = "✓"
		statusStyle = styles.RenderSuccess
	}

	b.WriteString(statusStyle(fmt.Sprintf("%s %s Sync", statusIcon, session.Type)))
	b.WriteString("\n")

	// Session details
	b.WriteString(fmt.Sprintf("  Started:  %s\n", session.StartTime.Format("2006-01-02 15:04:05")))
	if !session.EndTime.IsZero() {
		b.WriteString(fmt.Sprintf("  Ended:    %s\n", session.EndTime.Format("2006-01-02 15:04:05")))
		duration := session.EndTime.Sub(session.StartTime)
		b.WriteString(fmt.Sprintf("  Duration: %s\n", formatDuration(duration)))
	} else {
		b.WriteString("  Status:   In progress\n")
	}
	b.WriteString(fmt.Sprintf("  Files:    %d\n", session.Transfers))

	return b.String()
}

// renderStats renders backup statistics
func (m LogViewerModel) renderStats() tea.Msg {
	stats, err := m.logManager.GetStats()
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString(styles.RenderTitle("Backup Statistics"))
	b.WriteString("\n\n")

	b.WriteString("📊 Overall Statistics\n")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Total files backed up:  %s\n", 
		styles.RenderHighlight(fmt.Sprintf("%d", stats.TotalFiles))))
	b.WriteString(fmt.Sprintf("Total size:             %s\n", 
		styles.RenderHighlight(formatBytes(stats.TotalSize))))

	if !stats.LastSync.IsZero() {
		b.WriteString(fmt.Sprintf("Last sync:              %s\n", 
			stats.LastSync.Format("2006-01-02 15:04:05")))
	} else {
		b.WriteString("Last sync:              Never\n")
	}

	if !stats.LastSuccess.IsZero() {
		b.WriteString(fmt.Sprintf("Last successful sync:   %s\n", 
			stats.LastSuccess.Format("2006-01-02 15:04:05")))
	} else {
		b.WriteString("Last successful sync:   Never\n")
	}

	if stats.SuccessRate > 0 {
		successColor := styles.RenderSuccess
		if stats.SuccessRate < 80 {
			successColor = styles.RenderWarning
		}
		if stats.SuccessRate < 50 {
			successColor = styles.RenderError
		}
		b.WriteString(fmt.Sprintf("Success rate:           %s\n", 
			successColor(fmt.Sprintf("%.1f%%", stats.SuccessRate))))
	}

	b.WriteString("\n")

	// Recent activity
	sessions, _ := m.logManager.GetSyncSessions()
	if len(sessions) > 0 {
		b.WriteString("📈 Recent Activity (Last 5 sessions)\n")
		b.WriteString(strings.Repeat("─", 50))
		b.WriteString("\n\n")

		start := len(sessions) - 5
		if start < 0 {
			start = 0
		}

		for i := len(sessions) - 1; i >= start; i-- {
			session := sessions[i]
			statusIcon := "✗"
			if session.Success {
				statusIcon = "✓"
			}
			
			b.WriteString(fmt.Sprintf("%s %s - %s (%d files)\n",
				statusIcon,
				session.StartTime.Format("2006-01-02 15:04"),
				session.Type,
				session.Transfers,
			))
		}
	}

	return b.String()
}

// formatRelativeTime formats a duration as relative time
func formatRelativeTime(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
