package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/scripts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestConfig(tmpDir string) *scripts.Config {
	return &scripts.Config{
		HomeDir:      tmpDir,
		Username:     "testuser",
		RclonePath:   "/usr/local/bin/rclone",
		SourceRemote: "b2-backup",
		SourceBucket: "source-bucket",
		DestRemote:   "scaleway-backup",
		DestBucket:   "dest-bucket",
		LogDir:       filepath.Join(tmpDir, "logs"),
		BinDir:       filepath.Join(tmpDir, "bin"),
	}
}

func TestNewGenerator(t *testing.T) {
	gen := scripts.NewGenerator()
	assert.NotNil(t, gen)
}

func TestCreateDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)

	require.NoError(t, err)

	// Verify directories were created
	assert.DirExists(t, config.BinDir)
	assert.DirExists(t, config.LogDir)
}

func TestCreateDirectoriesAlreadyExist(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	// Pre-create directories
	require.NoError(t, os.MkdirAll(config.BinDir, 0755))
	require.NoError(t, os.MkdirAll(config.LogDir, 0755))

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)

	// Should not fail if directories already exist
	assert.NoError(t, err)
}

func TestGenerateEngineScript(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)
	require.NoError(t, err)

	err = gen.GenerateEngineScript(config)
	require.NoError(t, err)

	// Verify script was created
	scriptPath := filepath.Join(config.BinDir, "run_rclone_sync.sh")
	assert.FileExists(t, scriptPath)

	// Read and verify content
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, config.RclonePath)
	assert.Contains(t, contentStr, config.SourceRemote)
	assert.Contains(t, contentStr, config.DestRemote)
	assert.Contains(t, contentStr, "#!/bin/zsh")

	// Verify executable permissions
	info, err := os.Stat(scriptPath)
	require.NoError(t, err)
	mode := info.Mode()
	assert.True(t, mode&0111 != 0, "Script should be executable")
}

func TestGenerateMonthlyScript(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)
	require.NoError(t, err)

	err = gen.GenerateMonthlyScript(config)
	require.NoError(t, err)

	// Verify script was created
	scriptPath := filepath.Join(config.BinDir, "monthly_backup.sh")
	assert.FileExists(t, scriptPath)

	// Read and verify content
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#!/bin/zsh")
	assert.Contains(t, contentStr, config.LogDir)
	assert.Contains(t, contentStr, "lockfile")
	assert.Contains(t, contentStr, "timestamp")
}

func TestGenerateManualScript(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)
	require.NoError(t, err)

	err = gen.GenerateManualScript(config)
	require.NoError(t, err)

	// Verify script was created
	scriptPath := filepath.Join(config.BinDir, "sync_now.sh")
	assert.FileExists(t, scriptPath)

	// Read and verify content
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#!/bin/zsh")
	assert.Contains(t, contentStr, "Manual Sync")
}

func TestGenerateShowTransfersScript(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)
	require.NoError(t, err)

	err = gen.GenerateShowTransfersScript(config)
	require.NoError(t, err)

	// Verify script was created
	scriptPath := filepath.Join(config.BinDir, "show_transfers.sh")
	assert.FileExists(t, scriptPath)

	// Read and verify content
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#!/bin/zsh")
	assert.Contains(t, contentStr, config.LogDir)
}

func TestGenerateAllScripts(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)
	require.NoError(t, err)

	err = gen.GenerateAllScripts(config)
	require.NoError(t, err)

	// Verify all scripts were created
	scripts := []string{
		"run_rclone_sync.sh",
		"monthly_backup.sh",
		"sync_now.sh",
		"show_transfers.sh",
	}

	for _, script := range scripts {
		scriptPath := filepath.Join(config.BinDir, script)
		assert.FileExists(t, scriptPath, "Script %s should exist", script)

		// Verify executable
		info, err := os.Stat(scriptPath)
		require.NoError(t, err)
		mode := info.Mode()
		assert.True(t, mode&0111 != 0, "Script %s should be executable", script)
	}
}

func TestMakeExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.sh")

	// Create a test file
	err := os.WriteFile(testFile, []byte("#!/bin/bash\necho test"), 0644)
	require.NoError(t, err)

	// Make it executable
	gen := scripts.NewGenerator()
	err = gen.MakeExecutable(testFile)
	require.NoError(t, err)

	// Verify permissions
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	mode := info.Mode()
	assert.True(t, mode&0111 != 0, "File should be executable")
}

func TestScriptsValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *scripts.Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &scripts.Config{
				HomeDir:      "/home/user",
				Username:     "user",
				RclonePath:   "/usr/local/bin/rclone",
				SourceRemote: "b2",
				SourceBucket: "bucket1",
				DestRemote:   "scaleway",
				DestBucket:   "bucket2",
				LogDir:       "/home/user/logs",
				BinDir:       "/home/user/bin",
			},
			wantError: false,
		},
		{
			name: "missing rclone path",
			config: &scripts.Config{
				HomeDir:      "/home/user",
				Username:     "user",
				RclonePath:   "",
				SourceRemote: "b2",
				SourceBucket: "bucket1",
				DestRemote:   "scaleway",
				DestBucket:   "bucket2",
				LogDir:       "/home/user/logs",
				BinDir:       "/home/user/bin",
			},
			wantError: true,
		},
		{
			name: "missing source remote",
			config: &scripts.Config{
				HomeDir:      "/home/user",
				Username:     "user",
				RclonePath:   "/usr/local/bin/rclone",
				SourceRemote: "",
				SourceBucket: "bucket1",
				DestRemote:   "scaleway",
				DestBucket:   "bucket2",
				LogDir:       "/home/user/logs",
				BinDir:       "/home/user/bin",
			},
			wantError: true,
		},
		{
			name: "missing dest bucket",
			config: &scripts.Config{
				HomeDir:      "/home/user",
				Username:     "user",
				RclonePath:   "/usr/local/bin/rclone",
				SourceRemote: "b2",
				SourceBucket: "bucket1",
				DestRemote:   "scaleway",
				DestBucket:   "",
				LogDir:       "/home/user/logs",
				BinDir:       "/home/user/bin",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scripts.ValidateConfig(tt.config)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScriptContentSubstitution(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)
	require.NoError(t, err)

	// Generate engine script
	err = gen.GenerateEngineScript(config)
	require.NoError(t, err)

	// Read content
	scriptPath := filepath.Join(config.BinDir, "run_rclone_sync.sh")
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify no template placeholders remain
	assert.NotContains(t, contentStr, "{{")
	assert.NotContains(t, contentStr, "}}")

	// Verify all config values are present
	assert.Contains(t, contentStr, config.RclonePath)
	assert.Contains(t, contentStr, config.SourceRemote)
	assert.Contains(t, contentStr, config.SourceBucket)
	assert.Contains(t, contentStr, config.DestRemote)
	assert.Contains(t, contentStr, config.DestBucket)
}

func TestScriptPaths(t *testing.T) {
	tmpDir := t.TempDir()
	config := createTestConfig(tmpDir)

	tests := []struct {
		name       string
		scriptName string
		generate   func(*scripts.Generator, *scripts.Config) error
	}{
		{
			name:       "engine script",
			scriptName: "run_rclone_sync.sh",
			generate:   (*scripts.Generator).GenerateEngineScript,
		},
		{
			name:       "monthly script",
			scriptName: "monthly_backup.sh",
			generate:   (*scripts.Generator).GenerateMonthlyScript,
		},
		{
			name:       "manual script",
			scriptName: "sync_now.sh",
			generate:   (*scripts.Generator).GenerateManualScript,
		},
		{
			name:       "show transfers script",
			scriptName: "show_transfers.sh",
			generate:   (*scripts.Generator).GenerateShowTransfersScript,
		},
	}

	gen := scripts.NewGenerator()
	err := gen.CreateDirectories(config)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.generate(gen, config)
			require.NoError(t, err)

			scriptPath := filepath.Join(config.BinDir, tt.scriptName)
			assert.FileExists(t, scriptPath)

			// Verify shebang
			content, err := os.ReadFile(scriptPath)
			require.NoError(t, err)
			lines := strings.Split(string(content), "\n")
			assert.True(t, len(lines) > 0, "Script should have content")
			assert.Contains(t, lines[0], "#!/bin/zsh", "First line should be shebang")
		})
	}
}
