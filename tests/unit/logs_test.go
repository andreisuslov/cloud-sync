package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreisuslov/cloud-sync/internal/logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestLogFile(t *testing.T, path string, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}

func TestLogsNewManager(t *testing.T) {
	logDir := "/tmp/logs"
	manager := logs.NewManager(logDir)

	assert.NotNil(t, manager)
	assert.Contains(t, manager.GetLogPath(), "rclone_backup.log")
}

func TestLogsNewManagerWithPath(t *testing.T) {
	logPath := "/custom/path/backup.log"
	manager := logs.NewManagerWithPath(logPath)

	assert.NotNil(t, manager)
	assert.Equal(t, logPath, manager.GetLogPath())
}

func TestLogExists(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	tests := []struct {
		name       string
		createFile bool
		want       bool
	}{
		{
			name:       "log file exists",
			createFile: true,
			want:       true,
		},
		{
			name:       "log file does not exist",
			createFile: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createFile {
				createTestLogFile(t, logPath, "test log")
				defer os.Remove(logPath)
			}

			manager := logs.NewManagerWithPath(logPath)
			got := manager.LogExists()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetTodaysTransfers(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	today := time.Now().Format("2006/01/02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006/01/02")

	logContent := today + ` 14:30:45 INFO  : file1.txt: Copied (new)
` + today + ` 14:31:00 INFO  : file2.txt: Copied (new)
` + yesterday + ` 12:00:00 INFO  : old_file.txt: Copied (new)
` + today + ` 14:32:15 INFO  : file3.txt: Copied (new)
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)
	transfers, err := manager.GetTodaysTransfers()

	require.NoError(t, err)
	assert.Len(t, transfers, 3, "Should have 3 transfers from today")

	// Verify filenames
	filenames := make([]string, len(transfers))
	for i, t := range transfers {
		filenames[i] = t.Filename
	}
	assert.Contains(t, filenames, "file1.txt")
	assert.Contains(t, filenames, "file2.txt")
	assert.Contains(t, filenames, "file3.txt")
	assert.NotContains(t, filenames, "old_file.txt")
}

func TestGetRecentTransfers(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logContent := `2024/11/01 10:00:00 INFO  : file1.txt: Copied (new)
2024/11/01 10:01:00 INFO  : file2.txt: Copied (new)
2024/11/01 10:02:00 INFO  : file3.txt: Copied (new)
2024/11/01 10:03:00 INFO  : file4.txt: Copied (new)
2024/11/01 10:04:00 INFO  : file5.txt: Copied (new)
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)

	tests := []struct {
		name  string
		count int
		want  int
	}{
		{
			name:  "get 3 recent transfers",
			count: 3,
			want:  3,
		},
		{
			name:  "get 10 recent transfers (more than available)",
			count: 10,
			want:  5,
		},
		{
			name:  "get 0 transfers",
			count: 0,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transfers, err := manager.GetRecentTransfers(tt.count)
			require.NoError(t, err)
			assert.Len(t, transfers, tt.want)
		})
	}
}

func TestGetAllTransfers(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logContent := `2024/11/01 10:00:00 INFO  : file1.txt: Copied (new)
2024/11/01 10:01:00 INFO  : file2.txt: Copied (new)
2024/11/01 10:02:00 ERROR : Some error occurred
2024/11/01 10:03:00 INFO  : file3.txt: Copied (new)
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)
	transfers, err := manager.GetAllTransfers()

	require.NoError(t, err)
	assert.Len(t, transfers, 3, "Should have 3 transfers (errors excluded)")
}

func TestGetSyncSessions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logContent := `2024/11/01 09:00:00 INFO  : Manual Sync Requested
2024/11/01 09:01:00 INFO  : file1.txt: Copied (new)
2024/11/01 09:02:00 INFO  : file2.txt: Copied (new)
2024/11/01 09:05:00 INFO  : Manual Sync Complete: Success

2024/11/01 12:00:00 INFO  : Automated Check Started
2024/11/01 12:01:00 INFO  : file3.txt: Copied (new)
2024/11/01 12:05:00 INFO  : Backup successful
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)
	sessions, err := manager.GetSyncSessions()

	require.NoError(t, err)
	assert.Len(t, sessions, 2, "Should have 2 sync sessions")

	// Verify first session
	assert.Equal(t, "Manual", sessions[0].Type)
	assert.True(t, sessions[0].Success)
	assert.Equal(t, 2, sessions[0].Transfers)

	// Verify second session
	assert.Equal(t, "Automated", sessions[1].Type)
	assert.True(t, sessions[1].Success)
	assert.Equal(t, 1, sessions[1].Transfers)
}

func TestGetStats(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logContent := `2024/11/01 09:00:00 INFO  : Manual Sync Requested
2024/11/01 09:01:00 INFO  : file1.txt: Copied (new)
2024/11/01 09:02:00 INFO  : file2.txt: Copied (new)
2024/11/01 09:05:00 INFO  : Manual Sync Complete: Success

2024/11/01 12:00:00 INFO  : Manual Sync Requested
2024/11/01 12:01:00 INFO  : file3.txt: Copied (new)
2024/11/01 12:05:00 INFO  : Manual Sync Complete: Failed
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)
	stats, err := manager.GetStats()

	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalFiles)
	assert.False(t, stats.LastSync.IsZero())
	assert.False(t, stats.LastSuccess.IsZero())
	assert.Equal(t, 50.0, stats.SuccessRate, "Should be 50% success rate (1 of 2)")
}

func TestTailLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logContent := `Line 1
Line 2
Line 3
Line 4
Line 5
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)

	tests := []struct {
		name  string
		lines int
		want  int
	}{
		{
			name:  "tail 3 lines",
			lines: 3,
			want:  3,
		},
		{
			name:  "tail 10 lines (more than available)",
			lines: 10,
			want:  5,
		},
		{
			name:  "tail 0 lines",
			lines: 0,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := manager.TailLog(tt.lines)
			require.NoError(t, err)
			assert.Len(t, lines, tt.want)
		})
	}
}

func TestTailLogLastLines(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logContent := `Line 1
Line 2
Line 3
Line 4
Line 5
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)
	lines, err := manager.TailLog(2)

	require.NoError(t, err)
	assert.Len(t, lines, 2)
	assert.Equal(t, "Line 4", lines[0])
	assert.Equal(t, "Line 5", lines[1])
}

func TestClearOldLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	now := time.Now()
	old := now.AddDate(0, 0, -10)
	recent := now.AddDate(0, 0, -1)

	logContent := old.Format("2006/01/02 15:04:05") + ` INFO  : old_file.txt: Copied (new)
` + recent.Format("2006/01/02 15:04:05") + ` INFO  : recent_file.txt: Copied (new)
` + now.Format("2006/01/02 15:04:05") + ` INFO  : new_file.txt: Copied (new)
`

	createTestLogFile(t, logPath, logContent)
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)

	// Clear logs older than 7 days
	err := manager.ClearOldLogs(7 * 24 * time.Hour)
	require.NoError(t, err)

	// Read the file and check content
	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Old entry should be removed
	assert.NotContains(t, contentStr, "old_file.txt")

	// Recent entries should remain
	assert.Contains(t, contentStr, "recent_file.txt")
	assert.Contains(t, contentStr, "new_file.txt")
}

func TestGetTodaysTransfersEmptyLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "empty.log")

	createTestLogFile(t, logPath, "")
	defer os.Remove(logPath)

	manager := logs.NewManagerWithPath(logPath)
	transfers, err := manager.GetTodaysTransfers()

	require.NoError(t, err)
	assert.Empty(t, transfers)
}

func TestGetTodaysTransfersNoLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "nonexistent.log")

	manager := logs.NewManagerWithPath(logPath)
	transfers, err := manager.GetTodaysTransfers()

	require.NoError(t, err)
	assert.Empty(t, transfers)
}
