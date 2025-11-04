package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Installer handles installation of required tools
type Installer struct {
	stdout   *os.File
	stderr   *os.File
	executor CommandExecutor
}

// CommandExecutor interface for executing commands (allows mocking)
type CommandExecutor interface {
	LookPath(file string) (string, error)
	Command(name string, arg ...string) *exec.Cmd
	RunCommand(cmd *exec.Cmd) error
}

// DefaultExecutor implements CommandExecutor using real exec calls
type DefaultExecutor struct{}

func (e *DefaultExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (e *DefaultExecutor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func (e *DefaultExecutor) RunCommand(cmd *exec.Cmd) error {
	return cmd.Run()
}

// NewInstaller creates a new installer
func NewInstaller() *Installer {
	return &Installer{
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		executor: &DefaultExecutor{},
	}
}

// NewInstallerWithExecutor creates an installer with a custom executor (for testing)
func NewInstallerWithExecutor(executor CommandExecutor) *Installer {
	return &Installer{
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		executor: executor,
	}
}

// CheckHomebrewInstalled checks if Homebrew is installed
func (i *Installer) CheckHomebrewInstalled() bool {
	_, err := i.executor.LookPath("brew")
	return err == nil
}

// InstallHomebrew installs Homebrew
func (i *Installer) InstallHomebrew() error {
	if i.CheckHomebrewInstalled() {
		return fmt.Errorf("homebrew is already installed")
	}

	// Use the official Homebrew installation script
	installScript := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

	cmd := i.executor.Command("bash", "-c", installScript)
	cmd.Stdout = i.stdout
	cmd.Stderr = i.stderr
	cmd.Stdin = os.Stdin

	if err := i.executor.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to install homebrew: %w", err)
	}

	return nil
}

// CheckRcloneInstalled checks if rclone is installed
func (i *Installer) CheckRcloneInstalled() bool {
	_, err := i.executor.LookPath("rclone")
	return err == nil
}

// InstallRclone installs rclone via Homebrew
func (i *Installer) InstallRclone() error {
	if !i.CheckHomebrewInstalled() {
		return fmt.Errorf("homebrew must be installed first")
	}

	if i.CheckRcloneInstalled() {
		return fmt.Errorf("rclone is already installed")
	}

	cmd := i.executor.Command("brew", "install", "rclone")
	cmd.Stdout = i.stdout
	cmd.Stderr = i.stderr

	if err := i.executor.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to install rclone: %w", err)
	}

	return nil
}

// GetRclonePath returns the full path to rclone binary
func (i *Installer) GetRclonePath() (string, error) {
	path, err := i.executor.LookPath("rclone")
	if err != nil {
		return "", fmt.Errorf("rclone not found in PATH: %w", err)
	}

	// Resolve any symlinks
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If we can't resolve symlinks, return the original path
		return path, nil
	}

	return resolvedPath, nil
}

// GetBrewPath returns the path to brew binary
func (i *Installer) GetBrewPath() (string, error) {
	path, err := i.executor.LookPath("brew")
	if err != nil {
		return "", fmt.Errorf("brew not found in PATH: %w", err)
	}
	return path, nil
}

// GetArchitecture returns the system architecture (arm64 or amd64)
func GetArchitecture() string {
	return runtime.GOARCH
}

// GetHomebrewPrefix returns the expected Homebrew prefix for the architecture
func GetHomebrewPrefix() string {
	arch := GetArchitecture()
	if arch == "arm64" {
		return "/opt/homebrew"
	}
	return "/usr/local"
}

// VerifyInstallation verifies that all required tools are installed
func (i *Installer) VerifyInstallation() error {
	var errors []string

	if !i.CheckHomebrewInstalled() {
		errors = append(errors, "homebrew is not installed")
	}

	if !i.CheckRcloneInstalled() {
		errors = append(errors, "rclone is not installed")
	}

	if len(errors) > 0 {
		return fmt.Errorf("installation verification failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// CheckRsyncInstalled checks if rsync is installed
func (i *Installer) CheckRsyncInstalled() bool {
	_, err := i.executor.LookPath("rsync")
	return err == nil
}

// GetRsyncVersion returns the rsync version string
func (i *Installer) GetRsyncVersion() (string, error) {
	if !i.CheckRsyncInstalled() {
		return "", fmt.Errorf("rsync is not installed")
	}

	cmd := i.executor.Command("rsync", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get rsync version: %w", err)
	}

	// Parse version from output (first line typically contains version)
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "", fmt.Errorf("could not parse rsync version")
}

// GetRsyncPath returns the full path to rsync binary
func (i *Installer) GetRsyncPath() (string, error) {
	path, err := i.executor.LookPath("rsync")
	if err != nil {
		return "", fmt.Errorf("rsync not found in PATH: %w", err)
	}

	// Resolve any symlinks
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If we can't resolve symlinks, return the original path
		return path, nil
	}

	return resolvedPath, nil
}

// InstallRsync installs rsync via Homebrew
func (i *Installer) InstallRsync() error {
	if !i.CheckHomebrewInstalled() {
		return fmt.Errorf("homebrew must be installed first")
	}

	cmd := i.executor.Command("brew", "install", "rsync")
	cmd.Stdout = i.stdout
	cmd.Stderr = i.stderr

	if err := i.executor.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to install rsync: %w", err)
	}

	return nil
}

// UpdateRsync updates rsync via Homebrew
func (i *Installer) UpdateRsync() error {
	if !i.CheckHomebrewInstalled() {
		return fmt.Errorf("homebrew must be installed first")
	}

	if !i.CheckRsyncInstalled() {
		return fmt.Errorf("rsync is not installed, use InstallRsync instead")
	}

	cmd := i.executor.Command("brew", "upgrade", "rsync")
	cmd.Stdout = i.stdout
	cmd.Stderr = i.stderr

	if err := i.executor.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to update rsync: %w", err)
	}

	return nil
}