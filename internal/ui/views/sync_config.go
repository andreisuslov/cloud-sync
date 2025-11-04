package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/config"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
)

// SyncConfigModel represents the sync configuration wizard
type SyncConfigModel struct {
	inputs        []textinput.Model
	focusIndex    int
	width         int
	height        int
	err           error
	complete      bool
	configManager *config.Manager
	syncConfig    config.SyncConfig
}

// NewSyncConfigModel creates a new sync configuration model
func NewSyncConfigModel(configManager *config.Manager) SyncConfigModel {
	m := SyncConfigModel{
		configManager: configManager,
		inputs:        make([]textinput.Model, 4),
	}
	
	m.initInputs()
	return m
}

// Init initializes the sync configuration wizard
func (m SyncConfigModel) Init() tea.Cmd {
	return m.inputs[0].Focus()
}

// Update handles messages for the sync configuration wizard
func (m SyncConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc", "q":
			if m.complete {
				return m, nil
			}
			return m, nil

		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = styles.FocusedStyle
					m.inputs[i].TextStyle = styles.FocusedStyle
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = styles.NoStyle
					m.inputs[i].TextStyle = styles.NoStyle
				}
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.complete {
				return m, nil
			}
			return m.handleSave()
		}
	}

	// Handle character input for text fields
	if m.focusIndex < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the sync configuration wizard
func (m SyncConfigModel) View() string {
	helper := NewViewHelper(m.width, m.height)
	var b strings.Builder

	b.WriteString(helper.RenderHeader("Configure Sync Pairs", "Set up source and destination"))

	if m.complete {
		b.WriteString(m.renderComplete())
	} else {
		b.WriteString(m.renderForm())
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(styles.RenderError(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	if m.complete {
		b.WriteString(helper.RenderFooter("Press Enter to continue • q: Back"))
	} else {
		b.WriteString(helper.RenderFooter("Tab: Next field • Enter: Save • q: Back"))
	}

	return b.String()
}

// renderForm renders the configuration form
func (m SyncConfigModel) renderForm() string {
	var b strings.Builder
	
	b.WriteString(styles.RenderInfo("Sync Configuration"))
	b.WriteString("\n\n")
	b.WriteString(styles.RenderMuted("Configure the source and destination for your backup"))
	b.WriteString("\n\n")
	
	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n\n")
		}
	}
	
	b.WriteString("\n\n")
	b.WriteString(styles.RenderMuted("Example: Source Remote: b2, Source Bucket: notes-photos-documents"))
	b.WriteString("\n")
	b.WriteString(styles.RenderMuted("         Dest Remote: sw, Dest Bucket: b2-backup"))
	
	return b.String()
}

// renderComplete renders the completion message
func (m SyncConfigModel) renderComplete() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(2, 4).
		Width(60)

	content := styles.RenderSuccess("✓ Sync configuration saved!\n\n")
	content += fmt.Sprintf("Source: %s:%s\n", m.syncConfig.SourceRemote, m.syncConfig.SourceBucket)
	content += fmt.Sprintf("Destination: %s:%s\n", m.syncConfig.DestRemote, m.syncConfig.DestBucket)

	return box.Render(content)
}

// initInputs initializes the input fields
func (m *SyncConfigModel) initInputs() {
	// Source Remote
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "b2"
	m.inputs[0].Focus()
	m.inputs[0].PromptStyle = styles.FocusedStyle
	m.inputs[0].TextStyle = styles.FocusedStyle
	m.inputs[0].CharLimit = 32
	m.inputs[0].Width = 50
	m.inputs[0].Prompt = "Source Remote Name: "

	// Source Bucket
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "notes-photos-documents"
	m.inputs[1].CharLimit = 100
	m.inputs[1].Width = 50
	m.inputs[1].Prompt = "Source Bucket: "

	// Destination Remote
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "sw"
	m.inputs[2].CharLimit = 32
	m.inputs[2].Width = 50
	m.inputs[2].Prompt = "Dest Remote Name: "

	// Destination Bucket
	m.inputs[3] = textinput.New()
	m.inputs[3].Placeholder = "b2-backup"
	m.inputs[3].CharLimit = 100
	m.inputs[3].Width = 50
	m.inputs[3].Prompt = "Dest Bucket: "
}

// handleSave validates and saves the sync configuration
func (m SyncConfigModel) handleSave() (tea.Model, tea.Cmd) {
	sourceRemote := strings.TrimSpace(m.inputs[0].Value())
	sourceBucket := strings.TrimSpace(m.inputs[1].Value())
	destRemote := strings.TrimSpace(m.inputs[2].Value())
	destBucket := strings.TrimSpace(m.inputs[3].Value())

	if sourceRemote == "" || sourceBucket == "" || destRemote == "" || destBucket == "" {
		m.err = fmt.Errorf("all fields are required")
		return m, nil
	}

	m.syncConfig = config.SyncConfig{
		SourceRemote: sourceRemote,
		SourceBucket: sourceBucket,
		DestRemote:   destRemote,
		DestBucket:   destBucket,
	}

	if err := m.configManager.UpdateSyncConfig(m.syncConfig); err != nil {
		m.err = err
		return m, nil
	}

	m.complete = true
	return m, nil
}
