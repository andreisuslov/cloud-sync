package views

import (
	"fmt"
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
	StepComplete
)

// InstallationModel represents the installation wizard state
type InstallationModel struct {
	installer   *installer.Installer
	currentStep InstallationStep
	spinner     spinner.Model
	homebrewOK  bool
	rcloneOK    bool
	installing  bool
	error       error
	complete    bool
	width       int
	height      int
}

// NewInstallationModel creates a new installation wizard model
func NewInstallationModel() InstallationModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	return InstallationModel{
		installer:   installer.NewInstaller(),
		currentStep: StepCheckHomebrew,
		spinner:     s,
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
func (m InstallationModel) Update(msg tea.Msg) (InstallationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case checkResult:
		return m.handleCheckResult(msg)

	case installResult:
		return m.handleInstallResult(msg)
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
	case StepComplete:
		content = styles.RenderSuccess("✓ All tools installed successfully!")
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
func (m InstallationModel) handleCheckResult(msg checkResult) (InstallationModel, tea.Cmd) {
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
			m.currentStep = StepComplete
			m.complete = true
			return m, nil
		} else {
			m.currentStep = StepInstallRclone
			m.installing = true
			return m, m.installRclone()
		}
	}

	return m, nil
}

// handleInstallResult handles the result of an install operation
func (m InstallationModel) handleInstallResult(msg installResult) (InstallationModel, tea.Cmd) {
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
		m.currentStep = StepComplete
		m.complete = true
		return m, nil
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
