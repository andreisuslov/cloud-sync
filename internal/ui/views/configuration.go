package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/rclone"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
)

// ConfigurationStep represents a step in the configuration wizard
type ConfigurationStep int

const (
	StepWelcome ConfigurationStep = iota
	StepConfigureRemotes
	StepListRemotes
	StepSelectSourceBucket
	StepSelectDestBucket
	StepConfirmConfig
	StepComplete
)

// ConfigurationModel represents the configuration wizard state
type ConfigurationModel struct {
	rclone       *rclone.Manager
	currentStep  ConfigurationStep
	spinner      spinner.Model
	textInput    textinput.Model
	bucketList   list.Model
	sourceRemote string
	sourceBucket string
	destRemote   string
	destBucket   string
	remotes      []rclone.Remote
	buckets      []rclone.Bucket
	loading      bool
	error        error
	complete     bool
	width        int
	height       int
}

// NewConfigurationModel creates a new configuration wizard model
func NewConfigurationModel(rclonePath string) ConfigurationModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()

	return ConfigurationModel{
		rclone:      rclone.NewManager(rclonePath),
		currentStep: StepWelcome,
		spinner:     s,
		textInput:   ti,
	}
}

// Init initializes the configuration wizard
func (m ConfigurationModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the configuration wizard
func (m ConfigurationModel) Update(msg tea.Msg) (ConfigurationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.handleEnter()
		case "ctrl+c", "q":
			if m.currentStep == StepWelcome || m.complete {
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case remotesLoaded:
		m.remotes = msg.remotes
		m.loading = false
		if len(msg.remotes) == 0 {
			m.error = fmt.Errorf("no remotes configured, please run 'rclone config' first")
		}
		return m, nil

	case bucketsLoaded:
		m.buckets = msg.buckets
		m.loading = false
		return m, nil

	case configError:
		m.error = msg.err
		m.loading = false
		return m, nil
	}

	// Update active component
	var cmd tea.Cmd
	if m.textInput.Focused() {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

// View renders the configuration wizard
func (m ConfigurationModel) View() string {
	helper := NewViewHelper(m.width, m.height)
	var b strings.Builder

	b.WriteString(helper.RenderHeader("Configuration Wizard", "Set up your backup remotes and buckets"))

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
		b.WriteString(styles.RenderSuccess("✓ Configuration complete!"))
		b.WriteString("\n")
		b.WriteString(m.renderConfigSummary())
		b.WriteString(helper.RenderFooter("Press Enter to continue • q: Back to menu"))
	} else if m.loading {
		b.WriteString(helper.RenderFooter("Please wait..."))
	} else {
		b.WriteString(helper.RenderFooter("Enter: Continue • q: Back to menu"))
	}

	return b.String()
}

// renderCurrentStep renders the current configuration step
func (m ConfigurationModel) renderCurrentStep() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 2).
		Width(70)

	var content string

	switch m.currentStep {
	case StepWelcome:
		content = `Welcome to the Configuration Wizard!

This wizard will help you:
1. Configure rclone remotes for source and destination
2. Select the source bucket to backup from
3. Select the destination bucket to backup to

Prerequisites:
• Backblaze B2 account with bucket created
• Scaleway Object Storage account with bucket created
• API keys for both services

Press Enter to begin configuration...`

	case StepConfigureRemotes:
		if m.loading {
			content = fmt.Sprintf("%s Loading remotes...", m.spinner.View())
		} else {
			content = `Rclone Remote Configuration

To configure remotes, run the following in your terminal:
  rclone config

Configure two remotes:
1. Source remote (e.g., Backblaze B2)
2. Destination remote (e.g., Scaleway Object Storage)

After configuration, press Enter to continue...`
		}

	case StepListRemotes:
		if len(m.remotes) == 0 {
			content = "No remotes found. Please configure remotes first."
		} else {
			var remotesList strings.Builder
			remotesList.WriteString("Available Remotes:\n\n")
			for i, remote := range m.remotes {
				remotesList.WriteString(fmt.Sprintf("%d. %s\n", i+1, remote.Name))
			}
			remotesList.WriteString("\nEnter source remote name: ")
			remotesList.WriteString(m.textInput.View())
			content = remotesList.String()
		}

	case StepSelectSourceBucket:
		if m.loading {
			content = fmt.Sprintf("%s Loading buckets from %s...", m.spinner.View(), m.sourceRemote)
		} else if len(m.buckets) == 0 {
			content = fmt.Sprintf("No buckets found on %s", m.sourceRemote)
		} else {
			var bucketsList strings.Builder
			bucketsList.WriteString(fmt.Sprintf("Buckets on %s:\n\n", m.sourceRemote))
			for i, bucket := range m.buckets {
				bucketsList.WriteString(fmt.Sprintf("%d. %s\n", i+1, bucket.Name))
			}
			bucketsList.WriteString("\nEnter source bucket name: ")
			bucketsList.WriteString(m.textInput.View())
			content = bucketsList.String()
		}

	case StepSelectDestBucket:
		var destInput strings.Builder
		destInput.WriteString("Enter destination remote name: ")
		destInput.WriteString(m.textInput.View())
		destInput.WriteString("\n\nThen enter destination bucket name.")
		content = destInput.String()

	case StepConfirmConfig:
		content = m.renderConfigSummary()
		content += "\n\nPress Enter to confirm, q to cancel"

	case StepComplete:
		content = styles.RenderSuccess("✓ Configuration saved successfully!")
	}

	return box.Render(content)
}

// renderConfigSummary renders a summary of the configuration
func (m ConfigurationModel) renderConfigSummary() string {
	return fmt.Sprintf(`Configuration Summary:

Source:
  Remote: %s
  Bucket: %s

Destination:
  Remote: %s
  Bucket: %s`,
		m.sourceRemote, m.sourceBucket,
		m.destRemote, m.destBucket)
}

// handleEnter handles the Enter key press
func (m ConfigurationModel) handleEnter() (ConfigurationModel, tea.Cmd) {
	switch m.currentStep {
	case StepWelcome:
		m.currentStep = StepConfigureRemotes
		m.loading = true
		return m, m.loadRemotes()

	case StepConfigureRemotes:
		m.currentStep = StepListRemotes
		return m, nil

	case StepListRemotes:
		m.sourceRemote = m.textInput.Value()
		if m.sourceRemote != "" {
			m.currentStep = StepSelectSourceBucket
			m.loading = true
			m.textInput.Reset()
			return m, m.loadBuckets(m.sourceRemote)
		}

	case StepSelectSourceBucket:
		m.sourceBucket = m.textInput.Value()
		if m.sourceBucket != "" {
			m.currentStep = StepSelectDestBucket
			m.textInput.Reset()
			return m, nil
		}

	case StepSelectDestBucket:
		if m.destRemote == "" {
			m.destRemote = m.textInput.Value()
			m.textInput.Reset()
			return m, nil
		} else {
			m.destBucket = m.textInput.Value()
			if m.destBucket != "" {
				m.currentStep = StepConfirmConfig
				return m, nil
			}
		}

	case StepConfirmConfig:
		m.currentStep = StepComplete
		m.complete = true
		return m, nil
	}

	return m, nil
}

// loadRemotes loads the list of configured remotes
func (m ConfigurationModel) loadRemotes() tea.Cmd {
	return func() tea.Msg {
		remotes, err := m.rclone.ListRemotes()
		if err != nil {
			return configError{err: err}
		}
		return remotesLoaded{remotes: remotes}
	}
}

// loadBuckets loads buckets for a remote
func (m ConfigurationModel) loadBuckets(remoteName string) tea.Cmd {
	return func() tea.Msg {
		buckets, err := m.rclone.ListBuckets(remoteName)
		if err != nil {
			return configError{err: err}
		}
		return bucketsLoaded{buckets: buckets}
	}
}

// Message types
type remotesLoaded struct {
	remotes []rclone.Remote
}

type bucketsLoaded struct {
	buckets []rclone.Bucket
}

type configError struct {
	err error
}

// GetConfiguration returns the configured values
func (m ConfigurationModel) GetConfiguration() (sourceRemote, sourceBucket, destRemote, destBucket string) {
	return m.sourceRemote, m.sourceBucket, m.destRemote, m.destBucket
}
