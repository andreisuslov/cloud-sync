package syncconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SyncPair represents a local folder to remote sync configuration
type SyncPair struct {
	Name         string `json:"name"`          // User-friendly name for this sync
	LocalPath    string `json:"local_path"`    // Local folder path
	RemoteName   string `json:"remote_name"`   // Rclone remote name
	RemotePath   string `json:"remote_path"`   // Path on remote (bucket/folder)
	Direction    string `json:"direction"`     // "upload", "download", or "bidirectional"
	Enabled      bool   `json:"enabled"`       // Whether this sync is active
}

// Config holds all sync configurations
type Config struct {
	SyncPairs []SyncPair `json:"sync_pairs"`
	Version   string     `json:"version"`
}

// Manager handles sync configuration operations
type Manager struct {
	configPath string
}

// NewManager creates a new sync configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// NewDefaultManager creates a manager with default config path
func NewDefaultManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configPath := filepath.Join(homeDir, ".config", "cloud-sync", "sync-config.json")
	return &Manager{
		configPath: configPath,
	}, nil
}

// GetConfigPath returns the configuration file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// ConfigExists checks if the config file exists
func (m *Manager) ConfigExists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}

// Load loads the sync configuration from file
func (m *Manager) Load() (*Config, error) {
	if !m.ConfigExists() {
		// Return empty config if file doesn't exist
		return &Config{
			SyncPairs: []SyncPair{},
			Version:   "1.0",
		}, nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the sync configuration to file
func (m *Manager) Save(config *Config) error {
	// Ensure directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set version if not set
	if config.Version == "" {
		config.Version = "1.0"
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddSyncPair adds a new sync pair to the configuration
func (m *Manager) AddSyncPair(pair SyncPair) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	// Validate the sync pair
	if err := ValidateSyncPair(&pair); err != nil {
		return err
	}

	// Check for duplicates
	for _, existing := range config.SyncPairs {
		if existing.Name == pair.Name {
			return fmt.Errorf("sync pair with name '%s' already exists", pair.Name)
		}
		if existing.LocalPath == pair.LocalPath {
			return fmt.Errorf("local path '%s' is already configured", pair.LocalPath)
		}
	}

	config.SyncPairs = append(config.SyncPairs, pair)
	return m.Save(config)
}

// RemoveSyncPair removes a sync pair by name
func (m *Manager) RemoveSyncPair(name string) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	found := false
	newPairs := make([]SyncPair, 0, len(config.SyncPairs))
	for _, pair := range config.SyncPairs {
		if pair.Name != name {
			newPairs = append(newPairs, pair)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("sync pair '%s' not found", name)
	}

	config.SyncPairs = newPairs
	return m.Save(config)
}

// UpdateSyncPair updates an existing sync pair
func (m *Manager) UpdateSyncPair(name string, updatedPair SyncPair) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	// Validate the updated sync pair
	if err := ValidateSyncPair(&updatedPair); err != nil {
		return err
	}

	found := false
	for i, pair := range config.SyncPairs {
		if pair.Name == name {
			config.SyncPairs[i] = updatedPair
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("sync pair '%s' not found", name)
	}

	return m.Save(config)
}

// GetSyncPair retrieves a sync pair by name
func (m *Manager) GetSyncPair(name string) (*SyncPair, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	for _, pair := range config.SyncPairs {
		if pair.Name == name {
			return &pair, nil
		}
	}

	return nil, fmt.Errorf("sync pair '%s' not found", name)
}

// ListSyncPairs returns all sync pairs
func (m *Manager) ListSyncPairs() ([]SyncPair, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	return config.SyncPairs, nil
}

// ListEnabledSyncPairs returns only enabled sync pairs
func (m *Manager) ListEnabledSyncPairs() ([]SyncPair, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	enabled := make([]SyncPair, 0)
	for _, pair := range config.SyncPairs {
		if pair.Enabled {
			enabled = append(enabled, pair)
		}
	}

	return enabled, nil
}

// ToggleEnabled toggles the enabled status of a sync pair
func (m *Manager) ToggleEnabled(name string) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	found := false
	for i, pair := range config.SyncPairs {
		if pair.Name == name {
			config.SyncPairs[i].Enabled = !pair.Enabled
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("sync pair '%s' not found", name)
	}

	return m.Save(config)
}

// ValidateSyncPair validates a sync pair configuration
func ValidateSyncPair(pair *SyncPair) error {
	if pair.Name == "" {
		return fmt.Errorf("sync pair name cannot be empty")
	}

	if pair.LocalPath == "" {
		return fmt.Errorf("local path cannot be empty")
	}

	// Expand home directory if needed
	if pair.LocalPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		pair.LocalPath = filepath.Join(homeDir, pair.LocalPath[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(pair.LocalPath)
	if err != nil {
		return fmt.Errorf("invalid local path: %w", err)
	}
	pair.LocalPath = absPath

	if pair.RemoteName == "" {
		return fmt.Errorf("remote name cannot be empty")
	}

	if pair.RemotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}

	// Validate direction
	validDirections := map[string]bool{
		"upload":        true,
		"download":      true,
		"bidirectional": true,
	}
	if !validDirections[pair.Direction] {
		return fmt.Errorf("invalid direction '%s', must be 'upload', 'download', or 'bidirectional'", pair.Direction)
	}

	return nil
}

// ValidateLocalPath checks if a local path exists and is accessible
func ValidateLocalPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if we can read the directory
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read directory: %w", err)
	}
	file.Close()

	return nil
}
