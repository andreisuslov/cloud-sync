package launchd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

// Manager handles LaunchAgent operations
type Manager struct {
	username string
	agentPath string
}

// Config holds LaunchAgent configuration
type Config struct {
	Label      string // com.{username}.rclonebackup
	ScriptPath string // Path to monthly_backup.sh
	Hour       int    // Hour to run (0-23)
	Minute     int    // Minute to run (0-59)
	RunAtLoad  bool   // Run immediately when loaded
}

// Status represents the status of a LaunchAgent
type Status struct {
	Loaded   bool
	Running  bool
	PID      int
	LastExit int
}

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>

	<key>ProgramArguments</key>
	<array>
		<string>/bin/zsh</string>
		<string>{{.ScriptPath}}</string>
	</array>

	<key>StartCalendarInterval</key>
	<dict>
		<key>Hour</key>
		<integer>{{.Hour}}</integer>
		<key>Minute</key>
		<integer>{{.Minute}}</integer>
	</dict>
{{if .RunAtLoad}}
	<key>RunAtLoad</key>
	<true/>
{{end}}
</dict>
</plist>
`

// NewManager creates a new LaunchAgent manager
func NewManager(username string) *Manager {
	homeDir, _ := os.UserHomeDir()
	agentPath := filepath.Join(homeDir, "Library", "LaunchAgents")

	return &Manager{
		username: username,
		agentPath: agentPath,
	}
}

// GetLabel returns the LaunchAgent label for the user
func (m *Manager) GetLabel() string {
	return fmt.Sprintf("com.%s.rclonebackup", m.username)
}

// GetPlistPath returns the full path to the plist file
func (m *Manager) GetPlistPath() string {
	return filepath.Join(m.agentPath, m.GetLabel()+".plist")
}

// GeneratePlist generates a LaunchAgent plist file
func (m *Manager) GeneratePlist(config *Config) error {
	// Ensure LaunchAgents directory exists
	if err := os.MkdirAll(m.agentPath, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Parse template
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse plist template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("failed to execute plist template: %w", err)
	}

	// Write plist file
	plistPath := m.GetPlistPath()
	if err := os.WriteFile(plistPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	return nil
}

// Load loads the LaunchAgent
func (m *Manager) Load() error {
	plistPath := m.GetPlistPath()
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("plist file does not exist: %s", plistPath)
	}

	cmd := exec.Command("launchctl", "load", plistPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to load agent: %w (output: %s)", err, string(output))
	}

	return nil
}

// Unload unloads the LaunchAgent
func (m *Manager) Unload() error {
	plistPath := m.GetPlistPath()
	cmd := exec.Command("launchctl", "unload", plistPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Unload might fail if not loaded, which is okay
		if !strings.Contains(string(output), "Could not find specified service") {
			return fmt.Errorf("failed to unload agent: %w (output: %s)", err, string(output))
		}
	}

	return nil
}

// Start starts the LaunchAgent manually
func (m *Manager) Start() error {
	label := m.GetLabel()
	cmd := exec.Command("launchctl", "start", label)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start agent: %w (output: %s)", err, string(output))
	}

	return nil
}

// Stop stops the LaunchAgent
func (m *Manager) Stop() error {
	label := m.GetLabel()
	cmd := exec.Command("launchctl", "stop", label)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop agent: %w (output: %s)", err, string(output))
	}

	return nil
}

// GetStatus gets the status of the LaunchAgent
func (m *Manager) GetStatus() (*Status, error) {
	label := m.GetLabel()
	cmd := exec.Command("launchctl", "list", label)
	output, err := cmd.CombinedOutput()

	status := &Status{
		Loaded: false,
		Running: false,
		PID: -1,
		LastExit: 0,
	}

	if err != nil {
		// Service not found means it's not loaded
		if strings.Contains(string(output), "Could not find service") {
			return status, nil
		}
		return nil, fmt.Errorf("failed to get status: %w (output: %s)", err, string(output))
	}

	status.Loaded = true

	// Parse output
	// Format: "PID" = 12345; "LastExitStatus" = 0; ...
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "\"PID\"") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				pidStr := strings.TrimSpace(strings.Trim(parts[1], ";"))
				if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
					status.PID = pid
					status.Running = true
				}
			}
		}
		if strings.HasPrefix(line, "\"LastExitStatus\"") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				exitStr := strings.TrimSpace(strings.Trim(parts[1], ";"))
				if exitCode, err := strconv.Atoi(exitStr); err == nil {
					status.LastExit = exitCode
				}
			}
		}
	}

	return status, nil
}

// IsLoaded checks if the LaunchAgent is loaded
func (m *Manager) IsLoaded() (bool, error) {
	status, err := m.GetStatus()
	if err != nil {
		return false, err
	}
	return status.Loaded, nil
}

// Remove removes the plist file
func (m *Manager) Remove() error {
	// First try to unload if loaded
	if loaded, _ := m.IsLoaded(); loaded {
		if err := m.Unload(); err != nil {
			return fmt.Errorf("failed to unload before removing: %w", err)
		}
	}

	// Remove the plist file
	plistPath := m.GetPlistPath()
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	return nil
}

// ValidateConfig validates LaunchAgent configuration
func ValidateConfig(config *Config) error {
	if config.Label == "" {
		return fmt.Errorf("Label is required")
	}
	if config.ScriptPath == "" {
		return fmt.Errorf("ScriptPath is required")
	}
	if config.Hour < 0 || config.Hour > 23 {
		return fmt.Errorf("Hour must be between 0 and 23")
	}
	if config.Minute < 0 || config.Minute > 59 {
		return fmt.Errorf("Minute must be between 0 and 59")
	}
	return nil
}
