package unit

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/installer"
)

// MockRcloneExecutor is a mock executor for testing rclone operations
type MockRcloneExecutor struct {
	lookPathFunc   func(file string) (string, error)
	commandFunc    func(name string, arg ...string) *exec.Cmd
	runCommandFunc func(cmd *exec.Cmd) error
}

func (m *MockRcloneExecutor) LookPath(file string) (string, error) {
	if m.lookPathFunc != nil {
		return m.lookPathFunc(file)
	}
	return "", exec.ErrNotFound
}

func (m *MockRcloneExecutor) Command(name string, arg ...string) *exec.Cmd {
	if m.commandFunc != nil {
		return m.commandFunc(name, arg...)
	}
	return exec.Command("echo", "mock")
}

func (m *MockRcloneExecutor) RunCommand(cmd *exec.Cmd) error {
	if m.runCommandFunc != nil {
		return m.runCommandFunc(cmd)
	}
	return nil
}

// TestGetRcloneVersion tests getting rclone version
func TestGetRcloneVersion(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "rclone" {
				return "/opt/homebrew/bin/rclone", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			cmd := exec.Command("echo", "rclone v1.65.0")
			return cmd
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	version, err := inst.GetRcloneVersion()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if version == "" {
		t.Error("Expected version string, got empty")
	}
}

// TestGetRcloneVersionNotInstalled tests version check when rclone is not installed
func TestGetRcloneVersionNotInstalled(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	_, err := inst.GetRcloneVersion()
	if err == nil {
		t.Error("Expected error when rclone is not installed")
	}
}

// TestInstallRcloneWithOutput tests rclone installation
func TestInstallRcloneWithOutput(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "Installing rclone...")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	output, err := inst.InstallRcloneWithOutput()
	if err != nil {
		t.Errorf("Expected successful installation, got error: %v", err)
	}
	if output == "" {
		t.Error("Expected output from installation")
	}
}

// TestInstallRcloneWithoutHomebrew tests that installation fails without Homebrew
func TestInstallRcloneWithoutHomebrew(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	_, err := inst.InstallRcloneWithOutput()
	if err == nil {
		t.Error("Expected error when Homebrew is not installed")
	}
	expectedMsg := "homebrew must be installed first"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestUpdateRcloneWithOutput tests rclone update
func TestUpdateRcloneWithOutput(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" || file == "rclone" {
				return "/opt/homebrew/bin/" + file, nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "Updating rclone...")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	output, err := inst.UpdateRcloneWithOutput()
	if err != nil {
		t.Errorf("Expected successful update, got error: %v", err)
	}
	if output == "" {
		t.Error("Expected output from update")
	}
}

// TestUpdateRcloneNotInstalled tests that update fails when rclone is not installed
func TestUpdateRcloneNotInstalled(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	_, err := inst.UpdateRcloneWithOutput()
	if err == nil {
		t.Error("Expected error when rclone is not installed")
	}
}

// TestListRcloneRemotes tests listing configured remotes
func TestListRcloneRemotes(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "rclone" {
				return "/opt/homebrew/bin/rclone", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "remote1:\nremote2:")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	remotes, err := inst.ListRcloneRemotes()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(remotes) == 0 {
		t.Error("Expected remotes to be listed")
	}
}

// TestListRcloneRemotesNotInstalled tests listing remotes when rclone is not installed
func TestListRcloneRemotesNotInstalled(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	_, err := inst.ListRcloneRemotes()
	if err == nil {
		t.Error("Expected error when rclone is not installed")
	}
}

// TestTestRcloneRemote tests remote connectivity check
func TestTestRcloneRemote(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "rclone" {
				return "/opt/homebrew/bin/rclone", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "Connected successfully")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	output, err := inst.TestRcloneRemote("myremote")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if output == "" {
		t.Error("Expected output from remote test")
	}
}

// TestTestRcloneRemoteNotInstalled tests remote test when rclone is not installed
func TestTestRcloneRemoteNotInstalled(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	_, err := inst.TestRcloneRemote("myremote")
	if err == nil {
		t.Error("Expected error when rclone is not installed")
	}
}

// TestTestRcloneRemoteFailure tests remote test failure
func TestTestRcloneRemoteFailure(t *testing.T) {
	mock := &MockRcloneExecutor{
		lookPathFunc: func(file string) (string, error) {
			if file == "rclone" {
				return "/opt/homebrew/bin/rclone", nil
			}
			return "", exec.ErrNotFound
		},
		commandFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("false")
		},
		runCommandFunc: func(cmd *exec.Cmd) error {
			return fmt.Errorf("connection failed")
		},
	}
	inst := installer.NewInstallerWithExecutor(mock)

	_, err := inst.TestRcloneRemote("badremote")
	if err == nil {
		t.Error("Expected error when remote test fails")
	}
}
