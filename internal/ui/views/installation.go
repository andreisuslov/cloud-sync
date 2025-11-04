package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/config"
	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/andreisuslov/cloud-sync/internal/launchd"
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
	StepConfigureB2Remote
	StepConfigureScalewayRemote
	StepConfigureSyncPairs
	StepConfigureLaunchAgent
	StepGenerateScripts
	InstallStepComplete
)

// InstallationModel represents the installation wizard state
type InstallationModel struct {
	installer        *installer.Installer
	configManager    *config.Manager
	launchdMgr       *launchd.Manager
	currentStep      InstallationStep
	spinner          spinner.Model
	homebrewOK       bool
	rcloneOK         bool
	directoriesOK    bool
	b2Configured     bool
	scalewayConfigured bool
	syncConfigured   bool
	launchConfigured bool
	scriptsGenerated bool
	installing       bool
	error            error
	complete         bool
	width            int
	height           int
	homeDir          string
	binDir           string
	logDir           string
	activeSubView    tea.Model
}

// NewInstallationModel creates a new installation wizard model
func NewInstallationModel() InstallationModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	homeDir, _ := os.UserHomeDir()
	configMgr, _ := config.NewManager()
	username := os.Getenv("USER")
	launchdMgr := launchd.NewManager(username)

	return InstallationModel{
		installer:     installer.NewInstaller(),
		configManager: configMgr,
		launchdMgr:    launchdMgr,
		currentStep:   StepCheckHomebrew,
		spinner:       s,
		homeDir:       homeDir,
		binDir:        filepath.Join(homeDir, "bin"),
		logDir:        filepath.Join(homeDir, "logs"),
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
		// If we have an active sub-view, delegate to it
		if m.activeSubView != nil {
			var cmd tea.Cmd
			m.activeSubView, cmd = m.activeSubView.Update(msg)
			
			// Check if sub-view is complete
			if msg.String() == "enter" {
				switch m.currentStep {
				case StepConfigureB2Remote:
					if subModel, ok := m.activeSubView.(RemoteConfigModel); ok && subModel.complete {
						m.b2Configured = true
						m.activeSubView = nil
						m.currentStep = StepConfigureScalewayRemote
						return m, m.initScalewayConfig()
					}
				case StepConfigureScalewayRemote:
					if subModel, ok := m.activeSubView.(RemoteConfigModel); ok && subModel.complete {
						m.scalewayConfigured = true
						m.activeSubView = nil
						m.currentStep = StepConfigureSyncPairs
						return m, m.initSyncConfig()
					}
				case StepConfigureSyncPairs:
					if subModel, ok := m.activeSubView.(SyncConfigModel); ok && subModel.complete {
						m.syncConfigured = true
						m.activeSubView = nil
						m.currentStep = StepConfigureLaunchAgent
						return m, m.initLaunchAgentConfig()
					}
				case StepConfigureLaunchAgent:
					if subModel, ok := m.activeSubView.(LaunchAgentConfigModel); ok && subModel.complete {
						m.launchConfigured = true
						m.activeSubView = nil
						m.currentStep = StepGenerateScripts
						m.installing = true
						return m, m.generateScripts()
					}
				}
			}
			return m, cmd
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
		
	case scriptResult:
		return m.handleScriptResult(msg)
	}

	return m, nil
}

// View renders the installation wizard
func (m InstallationModel) View() string {
	// If we have an active sub-view, render it
	if m.activeSubView != nil {
		return m.activeSubView.View()
	}

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
		{"Configure B2", m.b2Configured || m.currentStep > StepConfigureB2Remote},
		{"Configure Scaleway", m.scalewayConfigured || m.currentStep > StepConfigureScalewayRemote},
		{"Configure sync", m.syncConfigured || m.currentStep > StepConfigureSyncPairs},
		{"Configure schedule", m.launchConfigured || m.currentStep > StepConfigureLaunchAgent},
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
	case StepConfigureB2Remote:
		content = styles.RenderInfo("Configuring Backblaze B2...")
	case StepConfigureScalewayRemote:
		content = styles.RenderInfo("Configuring Scaleway...")
	case StepConfigureSyncPairs:
		content = styles.RenderInfo("Configuring sync pairs...")
	case StepConfigureLaunchAgent:
		content = styles.RenderInfo("Configuring LaunchAgent...")
	case StepGenerateScripts:
		if m.scriptsGenerated {
			content = styles.RenderSuccess("✓ Scripts generated")
		} else {
			content = fmt.Sprintf("%s Generating backup scripts...", m.spinner.View())
		}
	case InstallStepComplete:
		content = styles.RenderSuccess("✓ Installation complete!\n\nAll tools installed and configured.\nBackups will run automatically according to your schedule.")
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
	m.currentStep = StepConfigureB2Remote
	return m, m.initB2Config()
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

// initB2Config initializes the B2 configuration sub-view
func (m InstallationModel) initB2Config() tea.Cmd {
	return func() tea.Msg {
		m.activeSubView = NewRemoteConfigModel(m.configManager)
		return m.activeSubView.Init()
	}
}

// initScalewayConfig initializes the Scaleway configuration sub-view
func (m InstallationModel) initScalewayConfig() tea.Cmd {
	return func() tea.Msg {
		m.activeSubView = NewRemoteConfigModel(m.configManager)
		return m.activeSubView.Init()
	}
}

// initSyncConfig initializes the sync configuration sub-view
func (m InstallationModel) initSyncConfig() tea.Cmd {
	return func() tea.Msg {
		m.activeSubView = NewSyncConfigModel(m.configManager)
		return m.activeSubView.Init()
	}
}

// initLaunchAgentConfig initializes the LaunchAgent configuration sub-view
func (m InstallationModel) initLaunchAgentConfig() tea.Cmd {
	return func() tea.Msg {
		m.activeSubView = NewLaunchAgentConfigModel(m.configManager, m.launchdMgr)
		return m.activeSubView.Init()
	}
}

// generateScripts generates backup scripts using stored configuration
func (m InstallationModel) generateScripts() tea.Cmd {
	return func() tea.Msg {
		// Load configuration
		appConfig, err := m.configManager.Load()
		if err != nil {
			return scriptResult{err: fmt.Errorf("failed to load config: %w", err)}
		}

		// Build remote paths
		sourcePath := fmt.Sprintf("%s:%s", appConfig.SyncConfig.SourceRemote, appConfig.SyncConfig.SourceBucket)
		destPath := fmt.Sprintf("%s:%s", appConfig.SyncConfig.DestRemote, appConfig.SyncConfig.DestBucket)

		// Create script templates with actual values
		scripts := map[string]string{
			"run_rclone_sync.sh": fmt.Sprintf(`#!/bin/zsh
# Rclone sync engine

LOG_DIR="%s"
LOG_FILE="$LOG_DIR/rclone_backup.log"
RCLONE_BIN="%s"
RCLONE_CONFIG="%s"

echo "--- Rclone Sync Starting: $(date) ---" >> "$LOG_FILE"

$RCLONE_BIN sync "%s" "%s" \
    --fast-list \
    --config "$RCLONE_CONFIG" \
    --log-file "$LOG_FILE" \
    -v \
    -P

exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "--- Rclone Sync Finished Successfully: $(date) ---" >> "$LOG_FILE"
else
    echo "--- Rclone Sync FAILED (Code: $exit_code): $(date) ---" >> "$LOG_FILE"
fi

echo "" >> "$LOG_FILE"
exit $exit_code
`, appConfig.LogDir, appConfig.RclonePath, appConfig.RcloneConfig, sourcePath, destPath),

			"monthly_backup.sh": fmt.Sprintf(`#!/bin/zsh
# Monthly backup script

LOG_DIR="%s"
TIMESTAMP_FILE="$LOG_DIR/rclone_last_run_timestamp"
LOG_FILE="$LOG_DIR/rclone_backup.log"

mkdir -p "$LOG_DIR"
touch "$TIMESTAMP_FILE"

CURRENT_MONTH=$(date +%%Y-%%m)
LAST_RUN_MONTH=$(cat "$TIMESTAMP_FILE")

echo "--- Automated Check Started: $(date) ---" >> "$LOG_FILE"

if [ "$CURRENT_MONTH" = "$LAST_RUN_MONTH" ]; then
    echo "Backup for $CURRENT_MONTH has already run. Skipping." >> "$LOG_FILE"
    echo "--- Automated Check Finished: $(date) ---" >> "$LOG_FILE"
    exit 0
fi

echo "Starting monthly backup for $CURRENT_MONTH..." >> "$LOG_FILE"

%s/run_rclone_sync.sh

if [ $? -eq 0 ]; then
    echo "Backup successful. Marking $CURRENT_MONTH as complete." >> "$LOG_FILE"
    echo "$CURRENT_MONTH" > "$TIMESTAMP_FILE"
else
    echo "ERROR: Rclone sync failed. Will retry on next scheduled run." >> "$LOG_FILE"
fi

echo "--- Automated Check Finished: $(date) ---" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"
`, appConfig.LogDir, appConfig.BinDir),

			"sync_now.sh": fmt.Sprintf(`#!/bin/zsh
# Manual sync script

echo "--- Manual Sync Requested: $(date) ---"
%s/run_rclone_sync.sh
echo "--- Manual Sync Complete. ---"
`, appConfig.BinDir),
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

type scriptResult struct {
	err error
}
