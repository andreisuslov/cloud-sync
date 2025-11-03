package lockfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles lockfile operations
type Manager struct {
	lockfilePath string
}

// NewManager creates a new lockfile manager
func NewManager(logDir string) *Manager {
	return &Manager{
		lockfilePath: filepath.Join(logDir, "rclone_backup.lock"),
	}
}

// NewManagerWithPath creates a manager with a custom lockfile path
func NewManagerWithPath(lockfilePath string) *Manager {
	return &Manager{
		lockfilePath: lockfilePath,
	}
}

// Exists checks if the lockfile exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.lockfilePath)
	return err == nil
}

// Create creates a lockfile
func (m *Manager) Create() error {
	if m.Exists() {
		return fmt.Errorf("lockfile already exists at %s", m.lockfilePath)
	}

	// Create the lockfile with current timestamp
	file, err := os.Create(m.lockfilePath)
	if err != nil {
		return fmt.Errorf("failed to create lockfile: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Created: %s\n", time.Now().Format(time.RFC3339)))
	if err != nil {
		return fmt.Errorf("failed to write to lockfile: %w", err)
	}

	return nil
}

// Remove removes the lockfile
func (m *Manager) Remove() error {
	if !m.Exists() {
		return nil // Already removed or never existed
	}

	err := os.Remove(m.lockfilePath)
	if err != nil {
		return fmt.Errorf("failed to remove lockfile: %w", err)
	}

	return nil
}

// GetAge returns the age of the lockfile
func (m *Manager) GetAge() (time.Duration, error) {
	if !m.Exists() {
		return 0, fmt.Errorf("lockfile does not exist")
	}

	info, err := os.Stat(m.lockfilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat lockfile: %w", err)
	}

	return time.Since(info.ModTime()), nil
}

// IsStale checks if the lockfile is older than maxAge
func (m *Manager) IsStale(maxAge time.Duration) bool {
	age, err := m.GetAge()
	if err != nil {
		return false
	}

	return age > maxAge
}

// ForceRemove removes the lockfile without checking if it exists
func (m *Manager) ForceRemove() error {
	// Try to remove, ignore errors if file doesn't exist
	err := os.Remove(m.lockfilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to force remove lockfile: %w", err)
	}
	return nil
}

// GetPath returns the lockfile path
func (m *Manager) GetPath() string {
	return m.lockfilePath
}