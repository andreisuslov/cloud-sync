package unit

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/installer"
)

// TestInstallRsyncWithoutHomebrew tests that InstallRsync fails without Homebrew
func TestInstallRsyncWithoutHomebrew(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			// Simulate Homebrew not installed
			if file == "brew" {
				return "", exec.ErrNotFound
			}
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.InstallRsync()
	if err == nil {
		t.Error("Expected error when Homebrew is not installed")
	}

	expectedMsg := "homebrew must be installed first"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestInstallRsyncSuccess tests successful rsync installation
func TestInstallRsyncSuccess(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			// rsync not installed initially
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "mock install")
		},
		runCommandFunc: func(cmd *exec.Cmd) error {
			// Simulate successful installation
			return nil
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.InstallRsync()
	if err != nil {
		t.Errorf("Expected successful installation, got error: %v", err)
	}
}

// TestInstallRsyncFailure tests failed rsync installation
func TestInstallRsyncFailure(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("false") // Command that always fails
		},
		runCommandFunc: func(cmd *exec.Cmd) error {
			// Simulate installation failure
			return fmt.Errorf("brew install failed: exit status 1")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.InstallRsync()
	if err == nil {
		t.Error("Expected error when installation fails")
	}

	t.Logf("Got expected error: %v", err)
}

// TestUpdateRsyncWithoutHomebrew tests that UpdateRsync fails without Homebrew
func TestUpdateRsyncWithoutHomebrew(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "", exec.ErrNotFound
			}
			if file == "rsync" {
				return "/usr/bin/rsync", nil
			}
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.UpdateRsync()
	if err == nil {
		t.Error("Expected error when Homebrew is not installed")
	}

	expectedMsg := "homebrew must be installed first"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestUpdateRsyncNotInstalled tests that UpdateRsync fails if rsync is not installed
func TestUpdateRsyncNotInstalled(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			// rsync not installed
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.UpdateRsync()
	if err == nil {
		t.Error("Expected error when rsync is not installed")
	}

	expectedMsg := "rsync is not installed, use InstallRsync instead"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestUpdateRsyncSystemVersion tests that UpdateRsync fails for system rsync
func TestUpdateRsyncSystemVersion(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			if file == "rsync" {
				return "/usr/bin/rsync", nil // System rsync
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("false")
		},
		runCommandFunc: func(cmd *exec.Cmd) error {
			// brew list rsync fails for system version
			return fmt.Errorf("Error: No such keg: /opt/homebrew/Cellar/rsync")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.UpdateRsync()
	if err == nil {
		t.Error("Expected error when trying to update system rsync")
	}

	t.Logf("Got expected error: %v", err)
}

// TestIsRsyncInstalledViaHomebrew tests the Homebrew detection
func TestIsRsyncInstalledViaHomebrew(t *testing.T) {
	tests := []struct {
		name     string
		mock     *MockRsyncExecutor
		expected bool
	}{
		{
			name: "rsync via Homebrew",
			mock: &MockRsyncExecutor{
				lookPathFunc: func(file string) (string, error) {
					if file == "brew" {
						return "/opt/homebrew/bin/brew", nil
					}
					return "", exec.ErrNotFound
				},
				commandFunc: func(name string, arg ...string) *exec.Cmd {
					return exec.Command("true")
				},
				runCommandFunc: func(cmd *exec.Cmd) error {
					return nil // brew list rsync succeeds
				},
			},
			expected: true,
		},
		{
			name: "system rsync",
			mock: &MockRsyncExecutor{
				lookPathFunc: func(file string) (string, error) {
					if file == "brew" {
						return "/opt/homebrew/bin/brew", nil
					}
					return "", exec.ErrNotFound
				},
				commandFunc: func(name string, arg ...string) *exec.Cmd {
					return exec.Command("false")
				},
				runCommandFunc: func(cmd *exec.Cmd) error {
					return fmt.Errorf("not installed via brew")
				},
			},
			expected: false,
		},
		{
			name: "no Homebrew",
			mock: &MockRsyncExecutor{
				lookPathFunc: func(file string) (string, error) {
					return "", exec.ErrNotFound
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := installer.NewInstallerWithExecutor(tt.mock)
			result := inst.IsRsyncInstalledViaHomebrew()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestUpdateRsyncSuccess tests successful rsync update
func TestUpdateRsyncSuccess(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			if file == "rsync" {
				return "/opt/homebrew/bin/rsync", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "mock update")
		},
		runCommandFunc: func(cmd *exec.Cmd) error {
			// Simulate successful update
			return nil
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.UpdateRsync()
	if err != nil {
		t.Errorf("Expected successful update, got error: %v", err)
	}
}

// TestUpdateRsyncFailure tests failed rsync update
func TestUpdateRsyncFailure(t *testing.T) {
	mock := &MockRsyncExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			if file == "rsync" {
				return "/opt/homebrew/bin/rsync", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("false")
		},
		runCommandFunc: func(cmd *exec.Cmd) error {
			// Simulate update failure
			return fmt.Errorf("brew upgrade failed: exit status 1")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	err := inst.UpdateRsync()
	if err == nil {
		t.Error("Expected error when update fails")
	}

	t.Logf("Got expected error: %v", err)
}

// TestRealInstallOrUpdate tests the actual install/update on the system
// This test is skipped by default but can be run manually
func TestRealInstallOrUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real installation test in short mode")
	}

	inst := installer.NewInstaller()

	// Check if Homebrew is installed
	if !inst.CheckHomebrewInstalled() {
		t.Skip("Homebrew not installed, skipping real installation test")
	}

	// Check if rsync is installed
	if inst.CheckRsyncInstalled() {
		t.Log("rsync is already installed, testing update...")
		err := inst.UpdateRsync()
		if err != nil {
			t.Logf("Update failed (this is expected if already up-to-date): %v", err)
		} else {
			t.Log("Update completed successfully")
		}
	} else {
		t.Log("rsync not installed, testing installation...")
		err := inst.InstallRsync()
		if err != nil {
			t.Errorf("Installation failed: %v", err)
		} else {
			t.Log("Installation completed successfully")
		}
	}
}
