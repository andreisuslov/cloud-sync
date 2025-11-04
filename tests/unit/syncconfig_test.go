package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/syncconfig"
)

func TestNewManager(t *testing.T) {
	configPath := "/tmp/test-sync-config.json"
	manager := syncconfig.NewManager(configPath)

	if manager.GetConfigPath() != configPath {
		t.Errorf("expected config path %s, got %s", configPath, manager.GetConfigPath())
	}
}

func TestNewDefaultManager(t *testing.T) {
	manager, err := syncconfig.NewDefaultManager()
	if err != nil {
		t.Fatalf("failed to create default manager: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".config", "cloud-sync", "sync-config.json")
	
	if manager.GetConfigPath() != expectedPath {
		t.Errorf("expected config path %s, got %s", expectedPath, manager.GetConfigPath())
	}
}

func TestConfigExists(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")
	manager := syncconfig.NewManager(configPath)

	// Should not exist initially
	if manager.ConfigExists() {
		t.Error("config should not exist initially")
	}

	// Create the file
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Should exist now
	if !manager.ConfigExists() {
		t.Error("config should exist after creation")
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")
	manager := syncconfig.NewManager(configPath)

	// Load non-existent config should return empty config
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(config.SyncPairs) != 0 {
		t.Error("expected empty sync pairs")
	}

	// Add a sync pair
	config.SyncPairs = append(config.SyncPairs, syncconfig.SyncPair{
		Name:       "test-sync",
		LocalPath:  "/tmp/test",
		RemoteName: "remote1",
		RemotePath: "bucket/folder",
		Direction:  "upload",
		Enabled:    true,
	})

	// Save config
	if err := manager.Save(config); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load again and verify
	loadedConfig, err := manager.Load()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if len(loadedConfig.SyncPairs) != 1 {
		t.Fatalf("expected 1 sync pair, got %d", len(loadedConfig.SyncPairs))
	}

	pair := loadedConfig.SyncPairs[0]
	if pair.Name != "test-sync" {
		t.Errorf("expected name 'test-sync', got '%s'", pair.Name)
	}
	if pair.LocalPath != "/tmp/test" {
		t.Errorf("expected local path '/tmp/test', got '%s'", pair.LocalPath)
	}
	if pair.RemoteName != "remote1" {
		t.Errorf("expected remote name 'remote1', got '%s'", pair.RemoteName)
	}
	if pair.Direction != "upload" {
		t.Errorf("expected direction 'upload', got '%s'", pair.Direction)
	}
	if !pair.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestAddSyncPair(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")
	manager := syncconfig.NewManager(configPath)

	// Create a test local directory
	localPath := filepath.Join(tmpDir, "local-folder")
	if err := os.MkdirAll(localPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	pair := syncconfig.SyncPair{
		Name:       "test-sync",
		LocalPath:  localPath,
		RemoteName: "remote1",
		RemotePath: "bucket/folder",
		Direction:  "upload",
		Enabled:    true,
	}

	// Add sync pair
	if err := manager.AddSyncPair(pair); err != nil {
		t.Fatalf("failed to add sync pair: %v", err)
	}

	// Verify it was added
	pairs, err := manager.ListSyncPairs()
	if err != nil {
		t.Fatalf("failed to list sync pairs: %v", err)
	}

	if len(pairs) != 1 {
		t.Fatalf("expected 1 sync pair, got %d", len(pairs))
	}

	// Try to add duplicate
	if err := manager.AddSyncPair(pair); err == nil {
		t.Error("expected error when adding duplicate sync pair")
	}
}

func TestRemoveSyncPair(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")
	manager := syncconfig.NewManager(configPath)

	// Create a test local directory
	localPath := filepath.Join(tmpDir, "local-folder")
	if err := os.MkdirAll(localPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	pair := syncconfig.SyncPair{
		Name:       "test-sync",
		LocalPath:  localPath,
		RemoteName: "remote1",
		RemotePath: "bucket/folder",
		Direction:  "upload",
		Enabled:    true,
	}

	// Add sync pair
	if err := manager.AddSyncPair(pair); err != nil {
		t.Fatalf("failed to add sync pair: %v", err)
	}

	// Remove it
	if err := manager.RemoveSyncPair("test-sync"); err != nil {
		t.Fatalf("failed to remove sync pair: %v", err)
	}

	// Verify it was removed
	pairs, err := manager.ListSyncPairs()
	if err != nil {
		t.Fatalf("failed to list sync pairs: %v", err)
	}

	if len(pairs) != 0 {
		t.Errorf("expected 0 sync pairs after removal, got %d", len(pairs))
	}

	// Try to remove non-existent
	if err := manager.RemoveSyncPair("non-existent"); err == nil {
		t.Error("expected error when removing non-existent sync pair")
	}
}

func TestUpdateSyncPair(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")
	manager := syncconfig.NewManager(configPath)

	// Create test local directories
	localPath1 := filepath.Join(tmpDir, "local-folder-1")
	localPath2 := filepath.Join(tmpDir, "local-folder-2")
	if err := os.MkdirAll(localPath1, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(localPath2, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	pair := syncconfig.SyncPair{
		Name:       "test-sync",
		LocalPath:  localPath1,
		RemoteName: "remote1",
		RemotePath: "bucket/folder",
		Direction:  "upload",
		Enabled:    true,
	}

	// Add sync pair
	if err := manager.AddSyncPair(pair); err != nil {
		t.Fatalf("failed to add sync pair: %v", err)
	}

	// Update it
	updatedPair := syncconfig.SyncPair{
		Name:       "test-sync",
		LocalPath:  localPath2,
		RemoteName: "remote2",
		RemotePath: "bucket2/folder2",
		Direction:  "download",
		Enabled:    false,
	}

	if err := manager.UpdateSyncPair("test-sync", updatedPair); err != nil {
		t.Fatalf("failed to update sync pair: %v", err)
	}

	// Verify update
	retrieved, err := manager.GetSyncPair("test-sync")
	if err != nil {
		t.Fatalf("failed to get sync pair: %v", err)
	}

	if retrieved.LocalPath != localPath2 {
		t.Errorf("expected local path '%s', got '%s'", localPath2, retrieved.LocalPath)
	}
	if retrieved.RemoteName != "remote2" {
		t.Errorf("expected remote name 'remote2', got '%s'", retrieved.RemoteName)
	}
	if retrieved.Direction != "download" {
		t.Errorf("expected direction 'download', got '%s'", retrieved.Direction)
	}
	if retrieved.Enabled {
		t.Error("expected enabled to be false")
	}
}

func TestToggleEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")
	manager := syncconfig.NewManager(configPath)

	// Create a test local directory
	localPath := filepath.Join(tmpDir, "local-folder")
	if err := os.MkdirAll(localPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	pair := syncconfig.SyncPair{
		Name:       "test-sync",
		LocalPath:  localPath,
		RemoteName: "remote1",
		RemotePath: "bucket/folder",
		Direction:  "upload",
		Enabled:    true,
	}

	// Add sync pair
	if err := manager.AddSyncPair(pair); err != nil {
		t.Fatalf("failed to add sync pair: %v", err)
	}

	// Toggle enabled
	if err := manager.ToggleEnabled("test-sync"); err != nil {
		t.Fatalf("failed to toggle enabled: %v", err)
	}

	// Verify it's disabled
	retrieved, err := manager.GetSyncPair("test-sync")
	if err != nil {
		t.Fatalf("failed to get sync pair: %v", err)
	}

	if retrieved.Enabled {
		t.Error("expected enabled to be false after toggle")
	}

	// Toggle again
	if err := manager.ToggleEnabled("test-sync"); err != nil {
		t.Fatalf("failed to toggle enabled: %v", err)
	}

	// Verify it's enabled
	retrieved, err = manager.GetSyncPair("test-sync")
	if err != nil {
		t.Fatalf("failed to get sync pair: %v", err)
	}

	if !retrieved.Enabled {
		t.Error("expected enabled to be true after second toggle")
	}
}

func TestListEnabledSyncPairs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")
	manager := syncconfig.NewManager(configPath)

	// Create test local directories
	localPath1 := filepath.Join(tmpDir, "local-folder-1")
	localPath2 := filepath.Join(tmpDir, "local-folder-2")
	if err := os.MkdirAll(localPath1, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(localPath2, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Add enabled pair
	pair1 := syncconfig.SyncPair{
		Name:       "enabled-sync",
		LocalPath:  localPath1,
		RemoteName: "remote1",
		RemotePath: "bucket/folder",
		Direction:  "upload",
		Enabled:    true,
	}
	if err := manager.AddSyncPair(pair1); err != nil {
		t.Fatalf("failed to add sync pair: %v", err)
	}

	// Add disabled pair
	pair2 := syncconfig.SyncPair{
		Name:       "disabled-sync",
		LocalPath:  localPath2,
		RemoteName: "remote2",
		RemotePath: "bucket2/folder2",
		Direction:  "download",
		Enabled:    false,
	}
	if err := manager.AddSyncPair(pair2); err != nil {
		t.Fatalf("failed to add sync pair: %v", err)
	}

	// List enabled pairs
	enabled, err := manager.ListEnabledSyncPairs()
	if err != nil {
		t.Fatalf("failed to list enabled sync pairs: %v", err)
	}

	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled sync pair, got %d", len(enabled))
	}

	if enabled[0].Name != "enabled-sync" {
		t.Errorf("expected enabled sync pair name 'enabled-sync', got '%s'", enabled[0].Name)
	}
}

func TestValidateSyncPair(t *testing.T) {
	tests := []struct {
		name      string
		pair      syncconfig.SyncPair
		shouldErr bool
	}{
		{
			name: "valid sync pair",
			pair: syncconfig.SyncPair{
				Name:       "test",
				LocalPath:  "/tmp",
				RemoteName: "remote1",
				RemotePath: "bucket/folder",
				Direction:  "upload",
			},
			shouldErr: false,
		},
		{
			name: "empty name",
			pair: syncconfig.SyncPair{
				LocalPath:  "/tmp",
				RemoteName: "remote1",
				RemotePath: "bucket/folder",
				Direction:  "upload",
			},
			shouldErr: true,
		},
		{
			name: "empty local path",
			pair: syncconfig.SyncPair{
				Name:       "test",
				RemoteName: "remote1",
				RemotePath: "bucket/folder",
				Direction:  "upload",
			},
			shouldErr: true,
		},
		{
			name: "empty remote name",
			pair: syncconfig.SyncPair{
				Name:       "test",
				LocalPath:  "/tmp",
				RemotePath: "bucket/folder",
				Direction:  "upload",
			},
			shouldErr: true,
		},
		{
			name: "empty remote path",
			pair: syncconfig.SyncPair{
				Name:       "test",
				LocalPath:  "/tmp",
				RemoteName: "remote1",
				Direction:  "upload",
			},
			shouldErr: true,
		},
		{
			name: "invalid direction",
			pair: syncconfig.SyncPair{
				Name:       "test",
				LocalPath:  "/tmp",
				RemoteName: "remote1",
				RemotePath: "bucket/folder",
				Direction:  "invalid",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := syncconfig.ValidateSyncPair(&tt.pair)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateLocalPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		path      string
		setup     func() error
		shouldErr bool
	}{
		{
			name: "valid directory",
			path: tmpDir,
			setup: func() error {
				return nil
			},
			shouldErr: false,
		},
		{
			name: "non-existent path",
			path: filepath.Join(tmpDir, "non-existent"),
			setup: func() error {
				return nil
			},
			shouldErr: true,
		},
		{
			name: "file instead of directory",
			path: filepath.Join(tmpDir, "file.txt"),
			setup: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			err := syncconfig.ValidateLocalPath(tt.path)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
