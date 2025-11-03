package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// BackupOperation represents the type of backup being performed
type BackupOperation int

const (
	BackupManual BackupOperation = iota
	BackupAutomated
)

// BackupStatus represents the current status of the backup
type BackupStatus int

const (
	BackupIdle BackupStatus = iota
	BackupRunning
	BackupCompleted
	BackupFailed
	BackupCancelled
)

// BackupProgress represents backup operation progress
type BackupProgress struct {
	CurrentFile   string
	FilesTotal    int
	FilesCopied   int
	BytesTotal    int64
	BytesCopied   int64
	Speed         float64 // MB/s
	StartTime     time.Time
	ElapsedTime   time.Duration
	ETA           time.Duration
	Status        BackupStatus
	ErrorMessage  string
}

// BackupOpsModel represents the backup operations view model
type BackupOpsModel struct {
	operation BackupOperation
	progress  BackupProgress
	spinner   spinner.Model
	progBar   progress.Model
	width     int
	height    int
	canceling bool
}

// NewBackupOpsModel creates a new backup operations model
func NewBackupOpsModel(operation BackupOperation, width, height int) BackupOpsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	p := progress.New(progress.WithDefaultGradient())

	return BackupOpsModel{
		operation: operation,
		progress: BackupProgress{
			Status:    BackupIdle,
			StartTime: time.Now(),
		},
		spinner:   s,
		progBar:   p,
		width:     width,
		height:    height,
		canceling: false,
	}
}

// Init implements tea.Model
func (m BackupOpsModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.startBackup())
}

// Update implements tea.Model
func (m BackupOpsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.progress.Status == BackupRunning {
				m.canceling = true
				return m, m.cancelBackup()
			}
			return m, tea.Quit
		case "enter":
			if m.progress.Status == BackupCompleted || 
			   m.progress.Status == BackupFailed || 
			   m.progress.Status == BackupCancelled {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progBar.Width = m.width - 4

	case BackupProgress:
		m.progress = msg
		if msg.Status == BackupRunning {
			return m, tea.Batch(m.spinner.Tick, m.tickProgress())
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progBar.Update(msg)
		m.progBar = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model
func (m BackupOpsModel) View() string {
	var b strings.Builder

	helper := NewViewHelper(m.width, m.height)

	// Header
	title := "Backup Operation"
	subtitle := ""
	switch m.operation {
	case BackupManual:
		subtitle = "Manual sync in progress"
	case BackupAutomated:
		subtitle = "Automated backup"
	}
	b.WriteString(helper.RenderHeader(title, subtitle))

	// Status
	switch m.progress.Status {
	case BackupIdle:
		b.WriteString(styles.RenderInfo("Preparing backup..."))
		b.WriteString("\n")
		b.WriteString(m.spinner.View())

	case BackupRunning:
		if m.canceling {
			b.WriteString(styles.RenderWarning("Cancelling backup..."))
			b.WriteString("\n")
		} else {
			// Progress bar
			percent := 0.0
			if m.progress.FilesTotal > 0 {
				percent = float64(m.progress.FilesCopied) / float64(m.progress.FilesTotal)
			}
			b.WriteString(m.progBar.ViewAs(percent))
			b.WriteString("\n\n")

			// Current file
			b.WriteString(styles.RenderInfo("Current file: "))
			if m.progress.CurrentFile != "" {
				b.WriteString(m.progress.CurrentFile)
			} else {
				b.WriteString("N/A")
			}
			b.WriteString("\n\n")

			// Statistics
			b.WriteString(m.renderStats())
			b.WriteString("\n")

			// Spinner
			b.WriteString(m.spinner.View())
			b.WriteString(" Working...")
		}

	case BackupCompleted:
		b.WriteString(styles.RenderSuccess("✓ Backup completed successfully!"))
		b.WriteString("\n\n")
		b.WriteString(m.renderSummary())

	case BackupFailed:
		b.WriteString(styles.RenderError("✗ Backup failed"))
		b.WriteString("\n\n")
		if m.progress.ErrorMessage != "" {
			b.WriteString(styles.RenderError("Error: " + m.progress.ErrorMessage))
			b.WriteString("\n\n")
		}
		b.WriteString(m.renderSummary())

	case BackupCancelled:
		b.WriteString(styles.RenderWarning("⚠ Backup cancelled"))
		b.WriteString("\n\n")
		b.WriteString(m.renderSummary())
	}

	// Footer
	helpText := ""
	if m.progress.Status == BackupRunning {
		helpText = "ctrl+c/q: Cancel backup"
	} else if m.progress.Status != BackupIdle {
		helpText = "enter: Return to menu • q: Quit"
	}
	b.WriteString(helper.RenderFooter(helpText))

	return b.String()
}

// renderStats renders backup statistics
func (m BackupOpsModel) renderStats() string {
	var b strings.Builder

	// Files
	b.WriteString(fmt.Sprintf("Files: %d / %d copied\n", 
		m.progress.FilesCopied, m.progress.FilesTotal))

	// Size
	sizeCopied := formatBytes(m.progress.BytesCopied)
	sizeTotal := formatBytes(m.progress.BytesTotal)
	b.WriteString(fmt.Sprintf("Size: %s / %s\n", sizeCopied, sizeTotal))

	// Speed
	b.WriteString(fmt.Sprintf("Speed: %.2f MB/s\n", m.progress.Speed))

	// Time
	m.progress.ElapsedTime = time.Since(m.progress.StartTime)
	b.WriteString(fmt.Sprintf("Elapsed: %s\n", formatDuration(m.progress.ElapsedTime)))

	// ETA
	if m.progress.ETA > 0 {
		b.WriteString(fmt.Sprintf("ETA: %s\n", formatDuration(m.progress.ETA)))
	}

	return b.String()
}

// renderSummary renders post-backup summary
func (m BackupOpsModel) renderSummary() string {
	var b strings.Builder

	b.WriteString("Summary:\n")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n")

	duration := m.progress.ElapsedTime
	if duration == 0 {
		duration = time.Since(m.progress.StartTime)
	}

	b.WriteString(fmt.Sprintf("Files copied: %d\n", m.progress.FilesCopied))
	b.WriteString(fmt.Sprintf("Total size: %s\n", formatBytes(m.progress.BytesCopied)))
	b.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(duration)))
	if m.progress.Speed > 0 {
		b.WriteString(fmt.Sprintf("Average speed: %.2f MB/s\n", m.progress.Speed))
	}

	return b.String()
}

// startBackup returns a command to start the backup operation
func (m BackupOpsModel) startBackup() tea.Cmd {
	return func() tea.Msg {
		// This would actually call the rclone sync
		// For now, return a mock progress update
		return BackupProgress{
			Status:      BackupRunning,
			FilesTotal:  100,
			FilesCopied: 0,
			BytesTotal:  1024 * 1024 * 1024, // 1GB
			BytesCopied: 0,
			StartTime:   time.Now(),
		}
	}
}

// tickProgress returns a command to update progress
func (m BackupOpsModel) tickProgress() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		// This would read actual progress from rclone
		// For now, simulate progress
		if m.progress.FilesCopied < m.progress.FilesTotal {
			return BackupProgress{
				Status:       BackupRunning,
				CurrentFile:  fmt.Sprintf("file_%d.dat", m.progress.FilesCopied+1),
				FilesTotal:   m.progress.FilesTotal,
				FilesCopied:  m.progress.FilesCopied + 1,
				BytesTotal:   m.progress.BytesTotal,
				BytesCopied:  m.progress.BytesCopied + (10 * 1024 * 1024),
				Speed:        10.5,
				StartTime:    m.progress.StartTime,
				ElapsedTime:  time.Since(m.progress.StartTime),
				ETA:          time.Duration(m.progress.FilesTotal-m.progress.FilesCopied) * time.Second,
			}
		}
		return BackupProgress{
			Status:       BackupCompleted,
			FilesTotal:   m.progress.FilesTotal,
			FilesCopied:  m.progress.FilesTotal,
			BytesTotal:   m.progress.BytesTotal,
			BytesCopied:  m.progress.BytesTotal,
			Speed:        m.progress.Speed,
			StartTime:    m.progress.StartTime,
			ElapsedTime:  time.Since(m.progress.StartTime),
		}
	})
}

// cancelBackup returns a command to cancel the backup
func (m BackupOpsModel) cancelBackup() tea.Cmd {
	return func() tea.Msg {
		// This would actually cancel the rclone process
		return BackupProgress{
			Status:       BackupCancelled,
			FilesTotal:   m.progress.FilesTotal,
			FilesCopied:  m.progress.FilesCopied,
			BytesTotal:   m.progress.BytesTotal,
			BytesCopied:  m.progress.BytesCopied,
			StartTime:    m.progress.StartTime,
			ElapsedTime:  time.Since(m.progress.StartTime),
		}
	}
}

// formatBytes formats bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration to human-readable format
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
