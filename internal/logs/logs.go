package logs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Manager handles log operations
type Manager struct {
	logFilePath string
}

// Transfer represents a file transfer entry
type Transfer struct {
	Timestamp time.Time
	Filename  string
	Size      int64
	Action    string // Copied, Deleted, etc.
}

// SyncSession represents a backup sync session
type SyncSession struct {
	StartTime time.Time
	EndTime   time.Time
	Success   bool
	Type      string // Manual, Automated
	Transfers int
}

// Stats represents backup statistics
type Stats struct {
	TotalFiles  int
	TotalSize   int64
	LastSync    time.Time
	LastSuccess time.Time
	SuccessRate float64
}

// NewManager creates a new log manager
func NewManager(logDir string) *Manager {
	return &Manager{
		logFilePath: filepath.Join(logDir, "rclone_backup.log"),
	}
}

// NewManagerWithPath creates a manager with custom log path
func NewManagerWithPath(logFilePath string) *Manager {
	return &Manager{
		logFilePath: logFilePath,
	}
}

// GetLogPath returns the log file path
func (m *Manager) GetLogPath() string {
	return m.logFilePath
}

// LogExists checks if the log file exists
func (m *Manager) LogExists() bool {
	_, err := os.Stat(m.logFilePath)
	return err == nil
}

// GetTodaysTransfers returns transfers from today
func (m *Manager) GetTodaysTransfers() ([]Transfer, error) {
	if !m.LogExists() {
		return []Transfer{}, nil
	}

	file, err := os.Open(m.logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	today := time.Now().Format("2006/01/02")
	var transfers []Transfer

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, today) && strings.Contains(line, "INFO") && strings.Contains(line, "Copied") {
			if transfer := parseTransferLine(line); transfer != nil {
				transfers = append(transfers, *transfer)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return transfers, nil
}

// GetRecentTransfers returns the most recent N transfers
func (m *Manager) GetRecentTransfers(count int) ([]Transfer, error) {
	if !m.LogExists() {
		return []Transfer{}, nil
	}

	file, err := os.Open(m.logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var allTransfers []Transfer

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "INFO") && strings.Contains(line, "Copied") {
			if transfer := parseTransferLine(line); transfer != nil {
				allTransfers = append(allTransfers, *transfer)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Return last N transfers
	if len(allTransfers) > count {
		return allTransfers[len(allTransfers)-count:], nil
	}

	return allTransfers, nil
}

// GetAllTransfers returns all transfers from the log
func (m *Manager) GetAllTransfers() ([]Transfer, error) {
	if !m.LogExists() {
		return []Transfer{}, nil
	}

	file, err := os.Open(m.logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var transfers []Transfer

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "INFO") && strings.Contains(line, "Copied") {
			if transfer := parseTransferLine(line); transfer != nil {
				transfers = append(transfers, *transfer)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return transfers, nil
}

// GetSyncSessions returns all sync sessions from the log
func (m *Manager) GetSyncSessions() ([]SyncSession, error) {
	if !m.LogExists() {
		return []SyncSession{}, nil
	}

	file, err := os.Open(m.logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var sessions []SyncSession
	var currentSession *SyncSession

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Detect session start
		if strings.Contains(line, "Manual Sync Requested") {
			if currentSession != nil {
				sessions = append(sessions, *currentSession)
			}
			currentSession = &SyncSession{
				StartTime: parseTimestamp(line),
				Type:      "Manual",
			}
		} else if strings.Contains(line, "Automated Check Started") {
			if currentSession != nil {
				sessions = append(sessions, *currentSession)
			}
			currentSession = &SyncSession{
				StartTime: parseTimestamp(line),
				Type:      "Automated",
			}
		}

		// Detect session end
		if currentSession != nil {
			if strings.Contains(line, "Manual Sync Complete: Success") || strings.Contains(line, "Backup successful") {
				currentSession.EndTime = parseTimestamp(line)
				currentSession.Success = true
			} else if strings.Contains(line, "Manual Sync Complete: Failed") || strings.Contains(line, "ERROR: Rclone sync failed") {
				currentSession.EndTime = parseTimestamp(line)
				currentSession.Success = false
			}

			// Count transfers in this session
			if strings.Contains(line, "INFO") && strings.Contains(line, "Copied") {
				currentSession.Transfers++
			}
		}
	}

	// Add last session if exists
	if currentSession != nil {
		sessions = append(sessions, *currentSession)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return sessions, nil
}

// GetStats returns backup statistics
func (m *Manager) GetStats() (*Stats, error) {
	sessions, err := m.GetSyncSessions()
	if err != nil {
		return nil, err
	}

	transfers, err := m.GetAllTransfers()
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		TotalFiles: len(transfers),
	}

	// Calculate total size
	for _, transfer := range transfers {
		stats.TotalSize += transfer.Size
	}

	// Find last sync and last success
	successCount := 0
	for i := len(sessions) - 1; i >= 0; i-- {
		session := sessions[i]
		if !session.EndTime.IsZero() && stats.LastSync.IsZero() {
			stats.LastSync = session.EndTime
		}
		if session.Success && stats.LastSuccess.IsZero() {
			stats.LastSuccess = session.EndTime
		}
		if session.Success {
			successCount++
		}
	}

	// Calculate success rate
	if len(sessions) > 0 {
		stats.SuccessRate = float64(successCount) / float64(len(sessions)) * 100
	}

	return stats, nil
}

// TailLog returns the last N lines from the log
func (m *Manager) TailLog(lines int) ([]string, error) {
	if !m.LogExists() {
		return []string{}, nil
	}

	file, err := os.Open(m.logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Return last N lines
	if len(allLines) > lines {
		return allLines[len(allLines)-lines:], nil
	}

	return allLines, nil
}

// parseTransferLine parses a log line containing transfer information
func parseTransferLine(line string) *Transfer {
	// Example: "2024/11/03 14:30:45 INFO  : file.txt: Copied (new)"
	re := regexp.MustCompile(`(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}).*INFO\s+:\s+(.+?):\s+Copied`)
	matches := re.FindStringSubmatch(line)

	if len(matches) < 3 {
		return nil
	}

	timestamp, err := time.Parse("2006/01/02 15:04:05", matches[1])
	if err != nil {
		return nil
	}

	return &Transfer{
		Timestamp: timestamp,
		Filename:  strings.TrimSpace(matches[2]),
		Action:    "Copied",
	}
}

// parseTimestamp extracts timestamp from a log line
func parseTimestamp(line string) time.Time {
	// Try to find a date pattern
	re := regexp.MustCompile(`(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})`)
	matches := re.FindStringSubmatch(line)

	if len(matches) > 1 {
		timestamp, err := time.Parse("2006/01/02 15:04:05", matches[1])
		if err == nil {
			return timestamp
		}
	}

	return time.Time{}
}

// ClearOldLogs removes log entries older than the specified duration
func (m *Manager) ClearOldLogs(olderThan time.Duration) error {
	if !m.LogExists() {
		return nil
	}

	file, err := os.Open(m.logFilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	cutoffTime := time.Now().Add(-olderThan)
	var newLines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		timestamp := parseTimestamp(line)

		// Keep lines without timestamps or recent lines
		if timestamp.IsZero() || timestamp.After(cutoffTime) {
			newLines = append(newLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	// Write back to file
	return os.WriteFile(m.logFilePath, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
}
