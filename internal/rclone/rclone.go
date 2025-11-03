package rclone

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles rclone operations
type Manager struct {
	configPath string
	rclonePath string
}

// Remote represents an rclone remote configuration
type Remote struct {
	Name     string
	Type     string
	Provider string
}

// Bucket represents a cloud storage bucket
type Bucket struct {
	Name         string
	Size         int64
	ModifiedTime time.Time
}

// NewManager creates a new rclone manager
func NewManager(rclonePath string) *Manager {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "rclone", "rclone.conf")

	return &Manager{
		configPath: configPath,
		rclonePath:  rclonePath,
	}
}

// NewManagerWithConfig creates a manager with custom config path
func NewManagerWithConfig(rclonePath, configPath string) *Manager {
	return &Manager{
		configPath: configPath,
		rclonePath:  rclonePath,
	}
}

// GetConfigPath returns the rclone config file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// ConfigExists checks if the config file exists
func (m *Manager) ConfigExists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}

// ListRemotes lists all configured remotes
func (m *Manager) ListRemotes() ([]Remote, error) {
	cmd := exec.Command(m.rclonePath, "listremotes", "--config", m.configPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	remotes := make([]Remote, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Remove trailing colon
		name := strings.TrimSuffix(line, ":")
		remotes = append(remotes, Remote{
			Name: name,
			// Type and Provider will be populated by GetRemoteInfo if needed
		})
	}

	return remotes, nil
}

// ListBuckets lists all buckets for a remote
func (m *Manager) ListBuckets(remoteName string) ([]Bucket, error) {
	cmd := exec.Command(m.rclonePath, "lsd", remoteName+":", "--config", m.configPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	buckets := make([]Bucket, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Parse output format: "-1 2023-01-01 12:00:00        -1 bucket-name"
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			bucketName := fields[len(fields)-1]
			buckets = append(buckets, Bucket{
				Name: bucketName,
			})
		}
	}

	return buckets, nil
}

// TestRemote tests connectivity to a remote
func (m *Manager) TestRemote(remoteName string) error {
	cmd := exec.Command(m.rclonePath, "lsd", remoteName+":", "--config", m.configPath, "--max-depth", "1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("remote test failed: %w", err)
	}
	return nil
}

// ConfigureRemote runs interactive rclone config
func (m *Manager) ConfigureRemote() error {
	cmd := exec.Command(m.rclonePath, "config", "--config", m.configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure remote: %w", err)
	}

	return nil
}

// ParseConfig parses the rclone config file
func (m *Manager) ParseConfig() (map[string]map[string]string, error) {
	if !m.ConfigExists() {
		return nil, fmt.Errorf("config file does not exist: %s", m.configPath)
	}

	file, err := os.Open(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := make(map[string]map[string]string)
	var currentSection string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			config[currentSection] = make(map[string]string)
			continue
		}

		// Key-value pair
		if currentSection != "" && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				config[currentSection][key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return config, nil
}

// GetRemoteType returns the type of a remote
func (m *Manager) GetRemoteType(remoteName string) (string, error) {
	config, err := m.ParseConfig()
	if err != nil {
		return "", err
	}

	remoteConfig, exists := config[remoteName]
	if !exists {
		return "", fmt.Errorf("remote %s not found in config", remoteName)
	}

	remoteType, exists := remoteConfig["type"]
	if !exists {
		return "", fmt.Errorf("type not found for remote %s", remoteName)
	}

	return remoteType, nil
}

// Sync performs a sync operation
func (m *Manager) Sync(source, dest string, progress bool, dryRun bool) error {
	args := []string{"sync", source, dest, "--config", m.configPath, "--fast-list", "-v"}

	if progress {
		args = append(args, "-P")
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.Command(m.rclonePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	return nil
}