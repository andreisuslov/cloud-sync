package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/syncconfig"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
)

// SyncPairsStep represents a step in the sync pairs wizard
type SyncPairsStep int

const (
	SyncPairsStepList SyncPairsStep = iota
	SyncPairsStepAdd
	SyncPairsStepAddName
	SyncPairsStepAddLocalPath
	SyncPairsStepAddRemoteName
	SyncPairsStepAddRemotePath
	SyncPairsStepAddDirection
	SyncPairsStepConfirm
	SyncPairsStepComplete
)

// SyncPairsModel represents the sync pairs management view
type SyncPairsModel struct {
	syncConfig  *syncconfig.Manager
	currentStep SyncPairsStep
	syncPairs   []syncconfig.SyncPair
	list        list.Model
	textInput   textinput.Model
	newPair     syncconfig.SyncPair
	error       error
	width       int
	height      int
	complete    bool
}

// NewSyncPairsModel creates a new sync pairs management model
func NewSyncPairsModel(syncConfigMgr *syncconfig.Manager) SyncPairsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()

	return SyncPairsModel{
		syncConfig:  syncConfigMgr,
		currentStep: SyncPairsStepList,
		textInput:   ti,
	}
}

// Init initializes the sync pairs view
func (m SyncPairsModel) Init() tea.Cmd {
	return m.loadSyncPairs()
}

// Update handles messages for the sync pairs view
func (m SyncPairsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.handleEnter()
		case "a":
			if m.currentStep == SyncPairsStepList {
				m.currentStep = SyncPairsStepAddName
				m.newPair = syncconfig.SyncPair{Enabled: true}
				m.textInput.Reset()
				return m, nil
			}
		case "d":
			if m.currentStep == SyncPairsStepList && len(m.syncPairs) > 0 {
				// Delete selected sync pair
				return m.handleDelete()
			}
		case "t":
			if m.currentStep == SyncPairsStepList && len(m.syncPairs) > 0 {
				// Toggle selected sync pair
				return m.handleToggle()
			}
		case "esc", "q":
			if m.currentStep == SyncPairsStepList || m.complete {
				return m, tea.Quit
			} else {
				// Go back to list
				m.currentStep = SyncPairsStepList
				m.error = nil
				return m, m.loadSyncPairs()
			}
		}

	case syncPairsLoaded:
		m.syncPairs = msg.pairs
		m.error = msg.err
		return m, nil
	}

	// Update active component
	var cmd tea.Cmd
	if m.textInput.Focused() {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

// View renders the sync pairs view
func (m SyncPairsModel) View() string {
	helper := NewViewHelper(m.width, m.height)
	var b strings.Builder

	b.WriteString(helper.RenderHeader("Sync Pairs Management", "Configure local folders to sync with cloud storage"))

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
		b.WriteString(styles.RenderSuccess("✓ Sync pair added successfully!"))
		b.WriteString("\n")
		b.WriteString(helper.RenderFooter("Press Enter to continue • q: Back to menu"))
	} else {
		b.WriteString(m.renderFooter())
	}

	return b.String()
}

// renderCurrentStep renders the current step
func (m SyncPairsModel) renderCurrentStep() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 2).
		Width(70)

	var content string

	switch m.currentStep {
	case SyncPairsStepList:
		content = m.renderSyncPairsList()

	case SyncPairsStepAddName:
		content = "Enter a name for this sync pair:\n\n"
		content += m.textInput.View()

	case SyncPairsStepAddLocalPath:
		content = "Enter the local folder path to sync:\n\n"
		content += m.textInput.View()
		content += "\n\nExample: /Users/username/Documents"

	case SyncPairsStepAddRemoteName:
		content = "Enter the rclone remote name:\n\n"
		content += m.textInput.View()
		content += "\n\nExample: backblaze, s3, gdrive"

	case SyncPairsStepAddRemotePath:
		content = "Enter the remote path (bucket/folder):\n\n"
		content += m.textInput.View()
		content += "\n\nExample: my-bucket/documents"

	case SyncPairsStepAddDirection:
		content = "Select sync direction:\n\n"
		content += "1. upload (local → remote)\n"
		content += "2. download (remote → local)\n"
		content += "3. bidirectional (both ways)\n\n"
		content += "Enter 1, 2, or 3: "
		content += m.textInput.View()

	case SyncPairsStepConfirm:
		content = m.renderNewPairSummary()
		content += "\n\nPress Enter to confirm, Esc to cancel"

	case SyncPairsStepComplete:
		content = styles.RenderSuccess("✓ Sync pair added successfully!")
	}

	return box.Render(content)
}

// renderSyncPairsList renders the list of sync pairs
func (m SyncPairsModel) renderSyncPairsList() string {
	if len(m.syncPairs) == 0 {
		return "No sync pairs configured.\n\nPress 'a' to add a new sync pair."
	}

	var b strings.Builder
	b.WriteString("Configured Sync Pairs:\n\n")

	for i, pair := range m.syncPairs {
		status := "✓"
		if !pair.Enabled {
			status = "✗"
		}

		b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, status, pair.Name))
		b.WriteString(fmt.Sprintf("   Local:  %s\n", pair.LocalPath))
		b.WriteString(fmt.Sprintf("   Remote: %s:%s\n", pair.RemoteName, pair.RemotePath))
		b.WriteString(fmt.Sprintf("   Direction: %s\n", pair.Direction))
		b.WriteString("\n")
	}

	return b.String()
}

// renderNewPairSummary renders a summary of the new sync pair
func (m SyncPairsModel) renderNewPairSummary() string {
	return fmt.Sprintf(`New Sync Pair Summary:

Name: %s
Local Path: %s
Remote: %s:%s
Direction: %s
Enabled: %v`,
		m.newPair.Name,
		m.newPair.LocalPath,
		m.newPair.RemoteName,
		m.newPair.RemotePath,
		m.newPair.Direction,
		m.newPair.Enabled)
}

// renderFooter renders the footer with available actions
func (m SyncPairsModel) renderFooter() string {
	helper := NewViewHelper(m.width, m.height)

	switch m.currentStep {
	case SyncPairsStepList:
		if len(m.syncPairs) > 0 {
			return helper.RenderFooter("a: Add • d: Delete • t: Toggle • q: Back")
		}
		return helper.RenderFooter("a: Add new sync pair • q: Back to menu")
	default:
		return helper.RenderFooter("Enter: Continue • Esc: Cancel • q: Back to menu")
	}
}

// handleEnter handles the Enter key press
func (m SyncPairsModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentStep {
	case SyncPairsStepAddName:
		m.newPair.Name = m.textInput.Value()
		if m.newPair.Name != "" {
			m.currentStep = SyncPairsStepAddLocalPath
			m.textInput.Reset()
		}

	case SyncPairsStepAddLocalPath:
		m.newPair.LocalPath = m.textInput.Value()
		if m.newPair.LocalPath != "" {
			m.currentStep = SyncPairsStepAddRemoteName
			m.textInput.Reset()
		}

	case SyncPairsStepAddRemoteName:
		m.newPair.RemoteName = m.textInput.Value()
		if m.newPair.RemoteName != "" {
			m.currentStep = SyncPairsStepAddRemotePath
			m.textInput.Reset()
		}

	case SyncPairsStepAddRemotePath:
		m.newPair.RemotePath = m.textInput.Value()
		if m.newPair.RemotePath != "" {
			m.currentStep = SyncPairsStepAddDirection
			m.textInput.Reset()
		}

	case SyncPairsStepAddDirection:
		choice := m.textInput.Value()
		switch choice {
		case "1":
			m.newPair.Direction = "upload"
		case "2":
			m.newPair.Direction = "download"
		case "3":
			m.newPair.Direction = "bidirectional"
		default:
			m.error = fmt.Errorf("invalid choice, please enter 1, 2, or 3")
			return m, nil
		}
		m.currentStep = SyncPairsStepConfirm
		m.textInput.Reset()

	case SyncPairsStepConfirm:
		// Add the sync pair
		if err := m.syncConfig.AddSyncPair(m.newPair); err != nil {
			m.error = err
			return m, nil
		}
		m.currentStep = SyncPairsStepComplete
		m.complete = true

	case SyncPairsStepComplete:
		m.currentStep = SyncPairsStepList
		m.complete = false
		m.error = nil
		return m, m.loadSyncPairs()
	}

	return m, nil
}

// handleDelete handles deleting a sync pair
func (m SyncPairsModel) handleDelete() (tea.Model, tea.Cmd) {
	// For simplicity, delete the first one
	// In a real implementation, you'd use a list selector
	if len(m.syncPairs) > 0 {
		if err := m.syncConfig.RemoveSyncPair(m.syncPairs[0].Name); err != nil {
			m.error = err
			return m, nil
		}
		return m, m.loadSyncPairs()
	}
	return m, nil
}

// handleToggle handles toggling a sync pair's enabled status
func (m SyncPairsModel) handleToggle() (tea.Model, tea.Cmd) {
	// For simplicity, toggle the first one
	// In a real implementation, you'd use a list selector
	if len(m.syncPairs) > 0 {
		if err := m.syncConfig.ToggleEnabled(m.syncPairs[0].Name); err != nil {
			m.error = err
			return m, nil
		}
		return m, m.loadSyncPairs()
	}
	return m, nil
}

// loadSyncPairs loads the list of sync pairs
func (m SyncPairsModel) loadSyncPairs() tea.Cmd {
	return func() tea.Msg {
		pairs, err := m.syncConfig.ListSyncPairs()
		return syncPairsLoaded{pairs: pairs, err: err}
	}
}

// Message types
type syncPairsLoaded struct {
	pairs []syncconfig.SyncPair
	err   error
}
