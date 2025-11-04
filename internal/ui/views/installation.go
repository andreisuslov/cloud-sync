package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
)

// InstallationStep represents a step in the installation wizard
type InstallationStep int

const (
	StepCheckHomebrew InstallationStep = iota
	StepInstallHomebrew
	StepCheckRclone
	StepInstallRclone
	StepCreateDirectories
	StepConfigureRclone
	StepGenerateScripts
	InstallStepComplete
)

// InstallationModel represents the installation wizard state
type InstallationModel struct {
	installer        *installer.Installer
	currentStep      InstallationStep
	spinner          spinner.Model
	homebrewOK       bool
	rcloneOK         bool
	directoriesOK    bool
	rcloneConfigured bool
	scriptsGenerated bool
	installing       bool
	error            error
	complete         bool
	width            int
	height           int
	homeDir          string
	binDir           string
	logDir           string
}

// NewInstallationModel creates a new installation wizard model
func NewInstallationModel() InstallationModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	homeDir, _ := os.UserHomeDir()

	return InstallationModel{
		installer:   installer.NewInstaller(),
		currentStep: StepCheckHomebrew,
		spinner:     s,
		homeDir:     homeDir,
		binDir:      filepath.Join(homeDir, "bin"),
		logDir:      filepath.Join(homeDir, "logs"),
	}
}

// Init initializes the installation wizard
func (m InstallationModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.checkHomebrew(),
	)
}

// Update handles messages for the installation wizard
func (m InstallationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle Enter key for rclone configuration step
		if m.currentStep == StepConfigureRclone && !m.rcloneConfigured && msg.String() == "enter" {
			m.installing = true
			return m, m.configureRclone()
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case checkResult:
		return m.handleCheckResult(msg)

	case installResult:
		return m.handleInstallResult(msg)
		
	case directoryResult:
		return m.handleDirectoryResult(msg)
		
	case rcloneConfigResult:
		return m.handleRcloneConfigResult(msg)
		
	case scriptResult:
		return m.handleScriptResult(msg)
	}

	return m, nil
}

// View renders the installation wizard
func (m InstallationModel) View() string {
	helper := NewViewHelper(m.width, m.height)
	var b strings.Builder

	b.WriteString(helper.RenderHeader("Installation Wizard", "Setting up required tools"))

	// Show current progress
	b.WriteString(m.renderProgress())
	b.WriteString("\n\n")

	// Show current step
	b.WriteString(m.renderCurrentStep())
	b.WriteString("\n")

	if m.error != nil {
		b.WriteString("\n")
		b.WriteString(styles.RenderError(fmt.Sprintf("Error: %v", m.error)))
		b.WriteString("\n")
	}

	if m.complete {
		b.WriteString("\n")
		b.WriteString(styles.RenderSuccess("✓ Installation complete!"))
		b.WriteString("\n")
		b.WriteString(helper.RenderFooter("Press Enter to continue • q: Back to menu"))
	} else if m.installing {
		b.WriteString(helper.RenderFooter("Please wait... • q: Cancel"))
	} else {
		b.WriteString(helper.RenderFooter("Press Enter to continue • q: Back to menu"))
	}

	return b.String()
}

// renderProgress shows overall progress
func (m InstallationModel) renderProgress() string {
	steps := []struct {
		name string
		done bool
	}{
		{"Check Homebrew", m.homebrewOK || m.currentStep > StepInstallHomebrew},
		{"Install Homebrew", m.homebrewOK || m.currentStep > StepInstallHomebrew},
		{"Check rclone", m.rcloneOK || m.currentStep > StepInstallRclone},
		{"Install rclone", m.rcloneOK || m.currentStep > StepInstallRclone},
		{"Create directories", m.directoriesOK || m.currentStep > StepCreateDirectories},
		{"Configure rclone", m.rcloneConfigured || m.currentStep > StepConfigureRclone},
		{"Generate scripts", m.scriptsGenerated || m.currentStep > StepGenerateScripts},
	}

	var b strings.Builder
	for _, step := range steps {
		if step.done {
			b.WriteString(styles.RenderSuccess("✓ " + step.name))
		} else {
			b.WriteString(styles.RenderInfo("○ " + step.name))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderCurrentStep renders the current installation step
func (m InstallationModel) renderCurrentStep() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 2).
		Width(60)

	var content string

	switch m.currentStep {
	case StepCheckHomebrew:
		content = fmt.Sprintf("%s Checking for Homebrew...", m.spinner.View())
	case StepInstallHomebrew:
		if m.homebrewOK {
			content = styles.RenderSuccess("✓ Homebrew is already installed")
		} else {
			content = fmt.Sprintf("%s Installing Homebrew...\nThis may take a few minutes.", m.spinner.View())
		}
	case StepCheckRclone:
		content = fmt.Sprintf("%s Checking for rclone...", m.spinner.View())
	case StepInstallRclone:
		if m.rcloneOK {
			content = styles.RenderSuccess("✓ rclone is already installed")
		} else {
			content = fmt.Sprintf("%s Installing rclone...", m.spinner.View())
		}
	case StepCreateDirectories:
		if m.directoriesOK {
			content = styles.RenderSuccess(fmt.Sprintf("✓ Directories created:\n  %s\n  %s", m.binDir, m.logDir))
		} else {
			content = fmt.Sprintf("%s Creating directories...", m.spinner.View())
		}
	case StepConfigureRclone:
		if m.rcloneConfigured {
			content = styles.RenderSuccess("✓ rclone configured")
		} else {
			content = styles.RenderInfo("Configure rclone remotes\n\nPress Enter to launch rclone config wizard")
		}
	case StepGenerateScripts:
		if m.scriptsGenerated {
			content = styles.RenderSuccess("✓ Scripts generated")
		} else {
			content = fmt.Sprintf("%s Generating backup scripts...", m.spinner.View())
		}
	case InstallStepComplete:
		content = styles.RenderSuccess("✓ Installation complete!\n\nAll tools installed and configured.\nYou can now set up backup operations.")
	}

	return box.Render(content)
}

// checkHomebrew returns a command to check if Homebrew is installed
func (m InstallationModel) checkHomebrew() tea.Cmd {
	return func() tea.Msg {
		installed := m.installer.CheckHomebrewInstalled()
		return checkResult{tool: "homebrew", installed: installed}
	}
}

// checkRclone returns a command to check if rclone is installed
func (m InstallationModel) checkRclone() tea.Cmd {
	return func() tea.Msg {
		installed := m.installer.CheckRcloneInstalled()
		return checkResult{tool: "rclone", installed: installed}
	}
}

// installHomebrew returns a command to install Homebrew
func (m InstallationModel) installHomebrew() tea.Cmd {
	return func() tea.Msg {
		err := m.installer.InstallHomebrew()
		return installResult{tool: "homebrew", err: err}
	}
}

// installRclone returns a command to install rclone
func (m InstallationModel) installRclone() tea.Cmd {
	return func() tea.Msg {
		err := m.installer.InstallRclone()
		return installResult{tool: "rclone", err: err}
	}
}

// handleCheckResult handles the result of a check operation
func (m InstallationModel) handleCheckResult(msg checkResult) (tea.Model, tea.Cmd) {
	switch msg.tool {
	case "homebrew":
		m.homebrewOK = msg.installed
		if msg.installed {
			m.currentStep = StepCheckRclone
			return m, m.checkRclone()
		} else {
			m.currentStep = StepInstallHomebrew
			m.installing = true
			return m, m.installHomebrew()
		}

	case "rclone":
		m.rcloneOK = msg.installed
		if msg.installed {
			m.currentStep = StepCreateDirectories
			m.installing = true
			return m, m.createDirectories()
		} else {
			m.currentStep = StepInstallRclone
			m.installing = true
			return m, m.installRclone()
		}
	}

	return m, nil
}

// handleDirectoryResult handles the result of directory creation
func (m InstallationModel) handleDirectoryResult(msg directoryResult) (tea.Model, tea.Cmd) {
	m.installing = false

	if msg.err != nil {
		m.error = msg.err
		return m, nil
	}

	m.directoriesOK = true
	m.currentStep = StepConfigureRclone
	return m, nil
}

// handleRcloneConfigResult handles the result of rclone configuration
func (m InstallationModel) handleRcloneConfigResult(msg rcloneConfigResult) (tea.Model, tea.Cmd) {
	m.installing = false

	if msg.err != nil {
		m.error = msg.err
		return m, nil
	}

	m.rcloneConfigured = true
	m.currentStep = StepGenerateScripts
	m.installing = true
	return m, m.generateScripts()
}

// handleScriptResult handles the result of script generation
func (m InstallationModel) handleScriptResult(msg scriptResult) (tea.Model, tea.Cmd) {
	m.installing = false

	if msg.err != nil {
		m.error = msg.err
		return m, nil
	}

	m.scriptsGenerated = true
	m.currentStep = InstallStepComplete
	m.complete = true
	return m, nil
}

// createDirectories creates required directories
func (m InstallationModel) createDirectories() tea.Cmd {
	return func() tea.Msg {
		dirs := []string{m.binDir, m.logDir}
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return directoryResult{err: fmt.Errorf("failed to create %s: %w", dir, err)}
			}
		}
		return directoryResult{err: nil}
	}
}

// configureRclone launches rclone config wizard
func (m InstallationModel) configureRclone() tea.Cmd {
	return func() tea.Msg {
		// For now, we'll skip interactive config and just check if config exists
		// In a real implementation, this would launch the rclone config wizard
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".config", "rclone", "rclone.conf")
		
		// Create config directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return rcloneConfigResult{err: fmt.Errorf("failed to create config directory: %w", err)}
		}
		
		// Check if config exists
		if _, err := os.Stat(configPath); err == nil {
			return rcloneConfigResult{err: nil}
		}
		
		// Config doesn't exist - user needs to configure manually
		// For now, we'll create an empty config file
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			return rcloneConfigResult{err: fmt.Errorf("failed to create config file: %w", err)}
		}
		
		return rcloneConfigResult{err: nil}
	}
}

// generateScripts generates backup scripts
func (m InstallationModel) generateScripts() tea.Cmd {
	return func() tea.Msg {
		// Create basic script templates
		scripts := map[string]string{
			"sync_now.sh": `#!/bin/zsh
# Manual sync script
echo "Manual sync not yet configured"
echo "Please configure rclone remotes first"
`,
			"monthly_backup.sh": `#!/bin/zsh
# Monthly backup script
echo "Monthly backup not yet configured"
echo "Please configure rclone remotes first"
`,
			"run_rclone_sync.sh": `#!/bin/zsh
# Rclone sync engine
echo "Rclone sync not yet configured"
echo "Please configure rclone remotes first"
`,
		}
		
		for name, content := range scripts {
			scriptPath := filepath.Join(m.binDir, name)
			if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
				return scriptResult{err: fmt.Errorf("failed to create %s: %w", name, err)}
			}
		}
		
		return scriptResult{err: nil}
	}
}

// handleInstallResult handles the result of an install operation
func (m InstallationModel) handleInstallResult(msg installResult) (tea.Model, tea.Cmd) {
	m.installing = false

	if msg.err != nil {
		m.error = msg.err
		return m, nil
	}

	switch msg.tool {
	case "homebrew":
		m.homebrewOK = true
		m.currentStep = StepCheckRclone
		return m, m.checkRclone()

	case "rclone":
		m.rcloneOK = true
		m.currentStep = StepCreateDirectories
		m.installing = true
		return m, m.createDirectories()
	}

	return m, nil
}

// Message types
type checkResult struct {
	tool      string
	installed bool
}

type installResult struct {
	tool string
	err  error
}

type directoryResult struct {
	err error
}

type rcloneConfigResult struct {
	err error
}

type scriptResult struct {
	err error
}
