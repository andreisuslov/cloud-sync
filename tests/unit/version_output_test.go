package unit

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/stretchr/testify/assert"
)

// TestGetRcloneVersionWithOutput tests the GetRcloneVersionWithOutput method
func TestGetRcloneVersionWithOutput(t *testing.T) {
	mockExec := new(MockExecutor)
	mockExec.On("LookPath", "rclone").Return("/usr/local/bin/rclone", nil)
	
	// Create a command that will output version info
	cmd := exec.Command("echo", "rclone v1.60.0")
	mockExec.On("Command", "rclone", []string{"version"}).Return(cmd)

	inst := installer.NewInstallerWithExecutor(mockExec)

	version, output, err := inst.GetRcloneVersionWithOutput()
	
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
	assert.NotEmpty(t, output)
	assert.Contains(t, version, "rclone")
	assert.Contains(t, output, "rclone")
	
	mockExec.AssertExpectations(t)
}

// TestGetRsyncVersionWithOutput tests the GetRsyncVersionWithOutput method
func TestGetRsyncVersionWithOutput(t *testing.T) {
	mockExec := new(MockExecutor)
	mockExec.On("LookPath", "rsync").Return("/usr/bin/rsync", nil)
	
	// Create a command that will output version info
	cmd := exec.Command("echo", "rsync  version 3.2.7  protocol version 31")
	mockExec.On("Command", "rsync", []string{"--version"}).Return(cmd)

	inst := installer.NewInstallerWithExecutor(mockExec)

	version, output, err := inst.GetRsyncVersionWithOutput()
	
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
	assert.NotEmpty(t, output)
	assert.Contains(t, version, "rsync")
	assert.Contains(t, output, "rsync")
	
	mockExec.AssertExpectations(t)
}

// TestGetRcloneVersionWithOutput_NotInstalled tests error handling when rclone is not installed
func TestGetRcloneVersionWithOutput_NotInstalled(t *testing.T) {
	mockExec := new(MockExecutor)
	mockExec.On("LookPath", "rclone").Return("", errors.New("not found"))

	inst := installer.NewInstallerWithExecutor(mockExec)

	_, _, err := inst.GetRcloneVersionWithOutput()
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
	
	mockExec.AssertExpectations(t)
}

// TestGetRsyncVersionWithOutput_NotInstalled tests error handling when rsync is not installed
func TestGetRsyncVersionWithOutput_NotInstalled(t *testing.T) {
	mockExec := new(MockExecutor)
	mockExec.On("LookPath", "rsync").Return("", errors.New("not found"))

	inst := installer.NewInstallerWithExecutor(mockExec)

	_, _, err := inst.GetRsyncVersionWithOutput()
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
	
	mockExec.AssertExpectations(t)
}
