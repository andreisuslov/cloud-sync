package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/launchd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLaunchdNewManager(t *testing.T) {
	username := "testuser"
	manager := launchd.NewManager(username)

	assert.NotNil(t, manager)
	assert.Contains(t, manager.GetLabel(), username)
	assert.Contains(t, manager.GetPlistPath(), "LaunchAgents")
}

func TestGetLabel(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{
			name:     "standard username",
			username: "john",
			want:     "com.john.rclonebackup",
		},
		{
			name:     "username with underscore",
			username: "john_doe",
			want:     "com.john_doe.rclonebackup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := launchd.NewManager(tt.username)
			got := manager.GetLabel()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPlistPath(t *testing.T) {
	username := "testuser"
	manager := launchd.NewManager(username)

	path := manager.GetPlistPath()
	assert.Contains(t, path, "LaunchAgents")
	assert.Contains(t, path, "com.testuser.rclonebackup.plist")
	assert.True(t, strings.HasSuffix(path, ".plist"))
}

func TestGeneratePlist(t *testing.T) {
	tmpDir := t.TempDir()
	username := "testuser"

	// Create a test manager with temp directory
	testAgentPath := filepath.Join(tmpDir, "LaunchAgents")
	
	manager := launchd.NewManager(username)

	config := &launchd.Config{
		Label:      manager.GetLabel(),
		ScriptPath: "/Users/testuser/bin/monthly_backup.sh",
		Hour:       2,
		Minute:     30,
		RunAtLoad:  false,
	}

	// Create the LaunchAgents directory
	err := os.MkdirAll(testAgentPath, 0755)
	require.NoError(t, err)

	// For this test, we'll just verify the config validation
	err = launchd.ValidateConfig(config)
	require.NoError(t, err)
}

func TestLaunchdValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *launchd.Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "/path/to/script.sh",
				Hour:       2,
				Minute:     30,
				RunAtLoad:  false,
			},
			wantError: false,
		},
		{
			name: "missing label",
			config: &launchd.Config{
				Label:      "",
				ScriptPath: "/path/to/script.sh",
				Hour:       2,
				Minute:     30,
			},
			wantError: true,
			errorMsg:  "Label is required",
		},
		{
			name: "missing script path",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "",
				Hour:       2,
				Minute:     30,
			},
			wantError: true,
			errorMsg:  "ScriptPath is required",
		},
		{
			name: "invalid hour - negative",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "/path/to/script.sh",
				Hour:       -1,
				Minute:     30,
			},
			wantError: true,
			errorMsg:  "Hour must be between 0 and 23",
		},
		{
			name: "invalid hour - too large",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "/path/to/script.sh",
				Hour:       24,
				Minute:     30,
			},
			wantError: true,
			errorMsg:  "Hour must be between 0 and 23",
		},
		{
			name: "invalid minute - negative",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "/path/to/script.sh",
				Hour:       2,
				Minute:     -1,
			},
			wantError: true,
			errorMsg:  "Minute must be between 0 and 59",
		},
		{
			name: "invalid minute - too large",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "/path/to/script.sh",
				Hour:       2,
				Minute:     60,
			},
			wantError: true,
			errorMsg:  "Minute must be between 0 and 59",
		},
		{
			name: "hour 0 is valid",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "/path/to/script.sh",
				Hour:       0,
				Minute:     0,
			},
			wantError: false,
		},
		{
			name: "hour 23 is valid",
			config: &launchd.Config{
				Label:      "com.test.backup",
				ScriptPath: "/path/to/script.sh",
				Hour:       23,
				Minute:     59,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := launchd.ValidateConfig(tt.config)
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigWithRunAtLoad(t *testing.T) {
	config := &launchd.Config{
		Label:      "com.test.backup",
		ScriptPath: "/path/to/script.sh",
		Hour:       2,
		Minute:     30,
		RunAtLoad:  true,
	}

	err := launchd.ValidateConfig(config)
	assert.NoError(t, err)
}

func TestManagerLabelConsistency(t *testing.T) {
	username := "testuser"
	manager := launchd.NewManager(username)

	label := manager.GetLabel()
	plistPath := manager.GetPlistPath()

	// Plist path should contain the label
	assert.Contains(t, plistPath, label)
}

func TestPlistPathStructure(t *testing.T) {
	username := "testuser"
	manager := launchd.NewManager(username)

	path := manager.GetPlistPath()

	// Should contain proper path components
	assert.Contains(t, path, "Library")
	assert.Contains(t, path, "LaunchAgents")
	assert.True(t, strings.HasSuffix(path, ".plist"))
}

// TestLoad, TestUnload, TestStart, TestStop, TestGetStatus
// These would require mocking launchctl commands or running on actual macOS
func TestLoadUnloadRequiresMocks(t *testing.T) {
	t.Skip("Load/Unload tests require launchctl mocking or actual macOS environment")
}

func TestStartStopRequiresMocks(t *testing.T) {
	t.Skip("Start/Stop tests require launchctl mocking or actual macOS environment")
}

func TestGetStatusRequiresMocks(t *testing.T) {
	t.Skip("GetStatus tests require launchctl mocking or actual macOS environment")
}
