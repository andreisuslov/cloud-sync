package unit

import (
	"os/exec"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/installer"
)

// TestCheckRsyncInstalled tests the rsync installation check
func TestCheckRsyncInstalled(t *testing.T) {
	inst := installer.NewInstaller()
	
	// This test assumes rsync is installed on macOS by default
	if !inst.CheckRsyncInstalled() {
		t.Error("Expected rsync to be installed on macOS")
	}
}

// TestGetRsyncVersion tests getting the rsync version
func TestGetRsyncVersion(t *testing.T) {
	inst := installer.NewInstaller()
	
	if !inst.CheckRsyncInstalled() {
		t.Skip("rsync is not installed, skipping version test")
	}
	
	version, err := inst.GetRsyncVersion()
	if err != nil {
		t.Errorf("Failed to get rsync version: %v", err)
	}
	
	if version == "" {
		t.Error("Expected non-empty version string")
	}
	
	t.Logf("rsync version: %s", version)
}

// TestGetRsyncPath tests getting the rsync path
func TestGetRsyncPath(t *testing.T) {
	inst := installer.NewInstaller()
	
	if !inst.CheckRsyncInstalled() {
		t.Skip("rsync is not installed, skipping path test")
	}
	
	path, err := inst.GetRsyncPath()
	if err != nil {
		t.Errorf("Failed to get rsync path: %v", err)
	}
	
	if path == "" {
		t.Error("Expected non-empty path string")
	}
	
	t.Logf("rsync path: %s", path)
}

// MockExecutor for testing installation functions
type MockRsyncExecutor struct {
	lookPathFunc   func(string) (string, error)
	commandFunc    func(string, ...string) *exec.Cmd
	runCommandFunc func(*exec.Cmd) error
}

func (m *MockRsyncExecutor) LookPath(file string) (string, error) {
	if m.lookPathFunc != nil {
		return m.lookPathFunc(file)
	}
	return "", exec.ErrNotFound
}

func (m *MockRsyncExecutor) Command(name string, arg ...string) *exec.Cmd {
	if m.commandFunc != nil {
		return m.commandFunc(name, arg...)
	}
	return exec.Command("echo", "mock")
}

func (m *MockRsyncExecutor) RunCommand(cmd *exec.Cmd) error {
	if m.runCommandFunc != nil {
		return m.runCommandFunc(cmd)
	}
	return nil
}

// TestCheckRsyncInstalledWithMock tests with a mock executor
func TestCheckRsyncInstalledWithMock(t *testing.T) {
	tests := []struct {
		name     string
		lookPath func(string) (string, error)
		expected bool
	}{
		{
			name: "rsync installed",
			lookPath: func(file string) (string, error) {
				if file == "rsync" {
					return "/usr/bin/rsync", nil
				}
				return "", exec.ErrNotFound
			},
			expected: true,
		},
		{
			name: "rsync not installed",
			lookPath: func(file string) (string, error) {
				return "", exec.ErrNotFound
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRsyncExecutor{
				lookPathFunc: tt.lookPath,
			}
			inst := installer.NewInstallerWithExecutor(mock)
			
			result := inst.CheckRsyncInstalled()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
