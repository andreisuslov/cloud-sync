package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RemoteConfig represents rclone remote configuration
type RemoteConfig struct {
	Name             string `json:"name"`
	Type             string `json:"type"`              // "b2" or "s3"
	Provider         string `json:"provider"`          // "Backblaze" or "Scaleway"
	AccountID        string `json:"account_id"`        // B2 account ID or S3 access key
	ApplicationKey   string `json:"application_key"`   // B2 app key or S3 secret key
	Region           string `json:"region,omitempty"`  // For S3
	Endpoint         string `json:"endpoint,omitempty"` // For S3
	Bucket           string `json:"bucket"`            // Default bucket for this remote
}

// SyncConfig represents sync operation configuration
type SyncConfig struct {
	SourceRemote string `json:"source_remote"` // Name of source remote
	SourceBucket string `json:"source_bucket"` // Source bucket/path
	DestRemote   string `json:"dest_remote"`   // Name of destination remote
	DestBucket   string `json:"dest_bucket"`   // Destination bucket/path
}

// LaunchAgentConfig represents LaunchAgent scheduling configuration
type LaunchAgentConfig struct {
	Enabled       bool   `json:"enabled"`
	Label         string `json:"label"`
	Hour          int    `json:"hour"`           // Hour to run (0-23)
	Minute        int    `json:"minute"`         // Minute to run (0-59)
	RunAtLoad     bool   `json:"run_at_load"`    // Run when loaded
	ScriptPath    string `json:"script_path"`    // Path to script to run
}

// AppConfig represents the complete application configuration
type AppConfig struct {
	Version       string              `json:"version"`
	Remotes       []RemoteConfig      `json:"remotes"`
	SyncConfig    SyncConfig          `json:"sync_config"`
	LaunchAgent   LaunchAgentConfig   `json:"launch_agent"`
	HomeDir       string              `json:"home_dir"`
	BinDir        string              `json:"bin_dir"`
	LogDir        string              `json:"log_dir"`
	RclonePath    string              `json:"rclone_path"`
	RcloneConfig  string              `json:"rclone_config"`
}

// Manager handles application configuration
type Manager struct {
	configPath string
	config     *AppConfig
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "cloud-sync")
	configPath := filepath.Join(configDir, "config.json")

	return &Manager{
		configPath: configPath,
	}, nil
}

// NewManagerWithPath creates a manager with custom config path
func NewManagerWithPath(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load loads the configuration from file
func (m *Manager) Load() (*AppConfig, error) {
	if !m.ConfigExists() {
		// Return default config if file doesn't exist
		return m.getDefaultConfig(), nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	m.config = &config
	return &config, nil
}

// Save saves the configuration to file
func (m *Manager) Save(config *AppConfig) error {
	// Ensure directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.config = config
	return nil
}

// ConfigExists checks if the config file exists
func (m *Manager) ConfigExists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}

// GetConfigPath returns the config file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// getDefaultConfig returns a default configuration
func (m *Manager) getDefaultConfig() *AppConfig {
	homeDir, _ := os.UserHomeDir()
	
	return &AppConfig{
		Version:      "1.0",
		Remotes:      []RemoteConfig{},
		SyncConfig:   SyncConfig{},
		LaunchAgent: LaunchAgentConfig{
			Enabled:   false,
			Label:     "com.cloud-sync.backup",
			Hour:      10,
			Minute:    5,
			RunAtLoad: true,
		},
		HomeDir:      homeDir,
		BinDir:       filepath.Join(homeDir, "bin"),
		LogDir:       filepath.Join(homeDir, "logs"),
		RclonePath:   "/opt/homebrew/bin/rclone",
		RcloneConfig: filepath.Join(homeDir, ".config", "rclone", "rclone.conf"),
	}
}

// AddRemote adds a new remote configuration
func (m *Manager) AddRemote(remote RemoteConfig) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	// Check for duplicate names
	for _, r := range config.Remotes {
		if r.Name == remote.Name {
			return fmt.Errorf("remote with name '%s' already exists", remote.Name)
		}
	}

	config.Remotes = append(config.Remotes, remote)
	return m.Save(config)
}

// UpdateRemote updates an existing remote configuration
func (m *Manager) UpdateRemote(name string, remote RemoteConfig) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	found := false
	for i, r := range config.Remotes {
		if r.Name == name {
			config.Remotes[i] = remote
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("remote '%s' not found", name)
	}

	return m.Save(config)
}

// RemoveRemote removes a remote configuration
func (m *Manager) RemoveRemote(name string) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	newRemotes := make([]RemoteConfig, 0)
	found := false
	for _, r := range config.Remotes {
		if r.Name != name {
			newRemotes = append(newRemotes, r)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("remote '%s' not found", name)
	}

	config.Remotes = newRemotes
	return m.Save(config)
}

// GetRemote retrieves a remote by name
func (m *Manager) GetRemote(name string) (*RemoteConfig, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	for _, r := range config.Remotes {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("remote '%s' not found", name)
}

// UpdateSyncConfig updates the sync configuration
func (m *Manager) UpdateSyncConfig(syncConfig SyncConfig) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	config.SyncConfig = syncConfig
	return m.Save(config)
}

// UpdateLaunchAgentConfig updates the LaunchAgent configuration
func (m *Manager) UpdateLaunchAgentConfig(launchConfig LaunchAgentConfig) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	config.LaunchAgent = launchConfig
	return m.Save(config)
}

// GenerateRcloneConfig generates rclone.conf from stored remotes
func (m *Manager) GenerateRcloneConfig() error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	// Ensure rclone config directory exists
	configDir := filepath.Dir(config.RcloneConfig)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create rclone config directory: %w", err)
	}

	// Generate rclone.conf content
	var content string
	for _, remote := range config.Remotes {
		content += fmt.Sprintf("[%s]\n", remote.Name)
		content += fmt.Sprintf("type = %s\n", remote.Type)
		
		if remote.Type == "b2" {
			content += fmt.Sprintf("account = %s\n", remote.AccountID)
			content += fmt.Sprintf("key = %s\n", remote.ApplicationKey)
		} else if remote.Type == "s3" {
			content += fmt.Sprintf("provider = %s\n", remote.Provider)
			content += fmt.Sprintf("access_key_id = %s\n", remote.AccountID)
			content += fmt.Sprintf("secret_access_key = %s\n", remote.ApplicationKey)
			if remote.Region != "" {
				content += fmt.Sprintf("region = %s\n", remote.Region)
			}
			if remote.Endpoint != "" {
				content += fmt.Sprintf("endpoint = %s\n", remote.Endpoint)
			}
		}
		content += "\n"
	}

	// Write to rclone.conf
	if err := os.WriteFile(config.RcloneConfig, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write rclone config: %w", err)
	}

	return nil
}
