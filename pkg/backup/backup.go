package backup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/andreisuslov/cloud-sync/internal/launchd"
	"github.com/andreisuslov/cloud-sync/internal/lockfile"
	"github.com/andreisuslov/cloud-sync/internal/logs"
	"github.com/andreisuslov/cloud-sync/internal/rclone"
	"github.com/andreisuslov/cloud-sync/internal/scripts"
	"github.com/andreisuslov/cloud-sync/internal/syncconfig"
)

// Manager provides a high-level API for backup operations
type Manager struct {
	installer  *installer.Installer
	rclone     *rclone.Manager
	scripts    *scripts.Generator
	launchd    *launchd.Manager
	logs       *logs.Manager
	lockfile   *lockfile.Manager
	syncconfig *syncconfig.Manager
	config     *Config
}

// Config holds the backup configuration
type Config struct {
	Username     string
	HomeDir      string
	SourceRemote string
	SourceBucket string
	DestRemote   string
	DestBucket   string
	RclonePath   string
	LogDir       string
	BinDir       string
}

// NewManager creates a new backup manager
func NewManager(config *Config) (*Manager, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults if not provided
	if config.HomeDir == "" {
		config.HomeDir, _ = os.UserHomeDir()
	}
	if config.LogDir == "" {
		config.LogDir = filepath.Join(config.HomeDir, "logs")
	}
	if config.BinDir == "" {
		config.BinDir = filepath.Join(config.HomeDir, "bin")
	}
	if config.RclonePath == "" {
		config.RclonePath = "rclone" // Will be resolved via PATH
	}

	// Initialize sync config manager
	syncConfigPath := filepath.Join(config.HomeDir, ".config", "cloud-sync", "sync-config.json")
	syncConfigMgr := syncconfig.NewManager(syncConfigPath)

	return &Manager{
		installer:  installer.NewInstaller(),
		rclone:     rclone.NewManager(config.RclonePath),
		scripts:    scripts.NewGenerator(),
		launchd:    launchd.NewManager(config.Username),
		logs:       logs.NewManager(config.LogDir),
		lockfile:   lockfile.NewManager(config.LogDir),
		syncconfig: syncConfigMgr,
		config:     config,
	}, nil
}

// VerifyPrerequisites checks if all required tools are installed
func (m *Manager) VerifyPrerequisites() error {
	return m.installer.VerifyInstallation()
}

// InstallTools installs Homebrew and rclone if needed
func (m *Manager) InstallTools() error {
	if !m.installer.CheckHomebrewInstalled() {
		if err := m.installer.InstallHomebrew(); err != nil {
			return fmt.Errorf("failed to install Homebrew: %w", err)
		}
	}

	if !m.installer.CheckRcloneInstalled() {
		if err := m.installer.InstallRclone(); err != nil {
			return fmt.Errorf("failed to install rclone: %w", err)
		}
	}

	// Update rclone path after installation
	path, err := m.installer.GetRclonePath()
	if err != nil {
		return fmt.Errorf("failed to get rclone path: %w", err)
	}
	m.config.RclonePath = path
	m.rclone = rclone.NewManager(path)

	return nil
}

// ConfigureRemotes runs interactive rclone configuration
func (m *Manager) ConfigureRemotes() error {
	return m.rclone.ConfigureRemote()
}

// ListRemotes lists all configured remotes
func (m *Manager) ListRemotes() ([]rclone.Remote, error) {
	return m.rclone.ListRemotes()
}

// ListBuckets lists buckets for a remote
func (m *Manager) ListBuckets(remoteName string) ([]rclone.Bucket, error) {
	return m.rclone.ListBuckets(remoteName)
}

// GenerateScripts generates all backup scripts
func (m *Manager) GenerateScripts() error {
	scriptConfig := &scripts.Config{
		HomeDir:      m.config.HomeDir,
		Username:     m.config.Username,
		RclonePath:   m.config.RclonePath,
		SourceRemote: m.config.SourceRemote,
		SourceBucket: m.config.SourceBucket,
		DestRemote:   m.config.DestRemote,
		DestBucket:   m.config.DestBucket,
		LogDir:       m.config.LogDir,
		BinDir:       m.config.BinDir,
	}

	if err := scripts.ValidateConfig(scriptConfig); err != nil {
		return fmt.Errorf("invalid script configuration: %w", err)
	}

	// Create directories
	if err := m.scripts.CreateDirectories(scriptConfig); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate all scripts
	if err := m.scripts.GenerateAllScripts(scriptConfig); err != nil {
		return fmt.Errorf("failed to generate scripts: %w", err)
	}

	return nil
}

// SetupLaunchAgent creates and loads the LaunchAgent
func (m *Manager) SetupLaunchAgent(hour, minute int) error {
	config := &launchd.Config{
		Label:      m.launchd.GetLabel(),
		ScriptPath: filepath.Join(m.config.BinDir, "monthly_backup.sh"),
		Hour:       hour,
		Minute:     minute,
		RunAtLoad:  false,
	}

	if err := launchd.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid LaunchAgent configuration: %w", err)
	}

	// Generate plist
	if err := m.launchd.GeneratePlist(config); err != nil {
		return fmt.Errorf("failed to generate plist: %w", err)
	}

	// Load the agent
	if err := m.launchd.Load(); err != nil {
		return fmt.Errorf("failed to load LaunchAgent: %w", err)
	}

	return nil
}

// GetLaunchAgentStatus returns the LaunchAgent status
func (m *Manager) GetLaunchAgentStatus() (*launchd.Status, error) {
	return m.launchd.GetStatus()
}

// StartManualBackup triggers a manual backup
func (m *Manager) StartManualBackup() error {
	if m.lockfile.Exists() {
		return fmt.Errorf("backup already running (lockfile exists)")
	}

	return m.launchd.Start()
}

// GetBackupStats returns backup statistics
func (m *Manager) GetBackupStats() (*logs.Stats, error) {
	return m.logs.GetStats()
}

// GetRecentTransfers returns recent transfers
func (m *Manager) GetRecentTransfers(count int) ([]logs.Transfer, error) {
	return m.logs.GetRecentTransfers(count)
}

// RemoveLockfile removes the backup lockfile
func (m *Manager) RemoveLockfile() error {
	return m.lockfile.ForceRemove()
}

// AddSyncPair adds a new local folder to remote sync configuration
func (m *Manager) AddSyncPair(name, localPath, remoteName, remotePath, direction string) error {
	pair := syncconfig.SyncPair{
		Name:       name,
		LocalPath:  localPath,
		RemoteName: remoteName,
		RemotePath: remotePath,
		Direction:  direction,
		Enabled:    true,
	}

	return m.syncconfig.AddSyncPair(pair)
}

// RemoveSyncPair removes a sync pair by name
func (m *Manager) RemoveSyncPair(name string) error {
	return m.syncconfig.RemoveSyncPair(name)
}

// ListSyncPairs returns all configured sync pairs
func (m *Manager) ListSyncPairs() ([]syncconfig.SyncPair, error) {
	return m.syncconfig.ListSyncPairs()
}

// ToggleSyncPair toggles the enabled status of a sync pair
func (m *Manager) ToggleSyncPair(name string) error {
	return m.syncconfig.ToggleEnabled(name)
}

// SyncPair executes a sync operation for a specific sync pair
func (m *Manager) SyncPair(name string, progress bool, dryRun bool) error {
	pair, err := m.syncconfig.GetSyncPair(name)
	if err != nil {
		return err
	}

	if !pair.Enabled {
		return fmt.Errorf("sync pair '%s' is disabled", name)
	}

	// Validate local path exists
	if err := syncconfig.ValidateLocalPath(pair.LocalPath); err != nil {
		return fmt.Errorf("local path validation failed: %w", err)
	}

	// Execute sync based on direction
	switch pair.Direction {
	case "upload":
		return m.rclone.SyncLocalToRemote(pair.LocalPath, pair.RemoteName, pair.RemotePath, progress, dryRun)
	case "download":
		return m.rclone.SyncRemoteToLocal(pair.RemoteName, pair.RemotePath, pair.LocalPath, progress, dryRun)
	case "bidirectional":
		// For bidirectional, we'll do upload first, then download
		// In a production system, you'd want more sophisticated conflict resolution
		if err := m.rclone.SyncLocalToRemote(pair.LocalPath, pair.RemoteName, pair.RemotePath, progress, dryRun); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		return m.rclone.SyncRemoteToLocal(pair.RemoteName, pair.RemotePath, pair.LocalPath, progress, dryRun)
	default:
		return fmt.Errorf("invalid sync direction: %s", pair.Direction)
	}
}

// SyncAllEnabled syncs all enabled sync pairs
func (m *Manager) SyncAllEnabled(progress bool, dryRun bool) error {
	pairs, err := m.syncconfig.ListEnabledSyncPairs()
	if err != nil {
		return err
	}

	if len(pairs) == 0 {
		return fmt.Errorf("no enabled sync pairs found")
	}

	for _, pair := range pairs {
		if err := m.SyncPair(pair.Name, progress, dryRun); err != nil {
			return fmt.Errorf("failed to sync '%s': %w", pair.Name, err)
		}
	}

	return nil
}

// validateConfig validates the manager configuration
func validateConfig(config *Config) error {
	if config.Username == "" {
		return fmt.Errorf("Username is required")
	}
	if config.SourceRemote == "" {
		return fmt.Errorf("SourceRemote is required")
	}
	if config.SourceBucket == "" {
		return fmt.Errorf("SourceBucket is required")
	}
	if config.DestRemote == "" {
		return fmt.Errorf("DestRemote is required")
	}
	if config.DestBucket == "" {
		return fmt.Errorf("DestBucket is required")
	}
	return nil
}
