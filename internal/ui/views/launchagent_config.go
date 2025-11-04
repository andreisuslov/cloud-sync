package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/config"
	"github.com/andreisuslov/cloud-sync/internal/launchd"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
)

// LaunchAgentConfigModel represents the LaunchAgent configuration wizard
type LaunchAgentConfigModel struct {
	inputs        []textinput.Model
	focusIndex    int
	width         int
	height        int
	err           error
	complete      bool
	configManager *config.Manager
	launchConfig  config.LaunchAgentConfig
	launchdMgr    *launchd.Manager
}

// NewLaunchAgentConfigModel creates a new LaunchAgent configuration model
func NewLaunchAgentConfigModel(configManager *config.Manager, launchdMgr *launchd.Manager) LaunchAgentConfigModel {
	m := LaunchAgentConfigModel{
		configManager: configManager,
		launchdMgr:    launchdMgr,
		inputs:        make([]textinput.Model, 2),
	}
	
	m.initInputs()
	return m
}

// Init initializes the LaunchAgent configuration wizard
func (m LaunchAgentConfigModel) Init() tea.Cmd {
	return m.inputs[0].Focus()
}

// Update handles messages for the LaunchAgent configuration wizard
func (m LaunchAgentConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// View renders the LaunchAgent configuration wizard
func (m LaunchAgentConfigModel) View() string {
	helper := NewViewHelper(m.width, m.height)
	var b strings.Builder

	b.WriteString(helper.RenderHeader("Configure LaunchAgent", "Set up automated backup scheduling"))

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
		b.WriteString(helper.RenderFooter("Tab: Next field • Enter: Save & Install • q: Back"))
	}

	return b.String()
}

// renderForm renders the configuration form
func (m LaunchAgentConfigModel) renderForm() string {
	var b strings.Builder
	
	b.WriteString(styles.RenderInfo("LaunchAgent Schedule Configuration"))
	b.WriteString("\n\n")
	b.WriteString(styles.RenderMuted("Configure when the automated backup should run"))
	b.WriteString("\n\n")
	
	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n\n")
		}
	}
	
	b.WriteString("\n\n")
	b.WriteString(styles.RenderMuted("The backup will run daily at the specified time"))
	b.WriteString("\n")
	b.WriteString(styles.RenderMuted("Example: Hour: 10, Minute: 5 = runs at 10:05 AM daily"))
	
	return b.String()
}

// renderComplete renders the completion message
func (m LaunchAgentConfigModel) renderComplete() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(2, 4).
		Width(60)

	content := styles.RenderSuccess("✓ LaunchAgent configured and installed!\n\n")
	content += fmt.Sprintf("Schedule: Daily at %02d:%02d\n", m.launchConfig.Hour, m.launchConfig.Minute)
	content += fmt.Sprintf("Label: %s\n", m.launchConfig.Label)
	content += "\nThe backup will run automatically according to this schedule."

	return box.Render(content)
}

// initInputs initializes the input fields
func (m *LaunchAgentConfigModel) initInputs() {
	// Hour
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "10"
	m.inputs[0].Focus()
	m.inputs[0].PromptStyle = styles.FocusedStyle
	m.inputs[0].TextStyle = styles.FocusedStyle
	m.inputs[0].CharLimit = 2
	m.inputs[0].Width = 50
	m.inputs[0].Prompt = "Hour (0-23): "
	m.inputs[0].SetValue("10")

	// Minute
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "5"
	m.inputs[1].CharLimit = 2
	m.inputs[1].Width = 50
	m.inputs[1].Prompt = "Minute (0-59): "
	m.inputs[1].SetValue("5")
}

// handleSave validates and saves the LaunchAgent configuration
func (m LaunchAgentConfigModel) handleSave() (tea.Model, tea.Cmd) {
	hourStr := strings.TrimSpace(m.inputs[0].Value())
	minuteStr := strings.TrimSpace(m.inputs[1].Value())

	if hourStr == "" || minuteStr == "" {
		m.err = fmt.Errorf("all fields are required")
		return m, nil
	}

	hour, err := strconv.Atoi(hourStr)
	if err != nil || hour < 0 || hour > 23 {
		m.err = fmt.Errorf("hour must be between 0 and 23")
		return m, nil
	}

	minute, err := strconv.Atoi(minuteStr)
	if err != nil || minute < 0 || minute > 59 {
		m.err = fmt.Errorf("minute must be between 0 and 59")
		return m, nil
	}

	// Load current config to get paths
	appConfig, err := m.configManager.Load()
	if err != nil {
		m.err = fmt.Errorf("failed to load config: %w", err)
		return m, nil
	}

	m.launchConfig = config.LaunchAgentConfig{
		Enabled:    true,
		Label:      "com.cloud-sync.backup",
		Hour:       hour,
		Minute:     minute,
		RunAtLoad:  true,
		ScriptPath: appConfig.BinDir + "/monthly_backup.sh",
	}

	// Save to config
	if err := m.configManager.UpdateLaunchAgentConfig(m.launchConfig); err != nil {
		m.err = fmt.Errorf("failed to save config: %w", err)
		return m, nil
	}

	// Generate plist file
	launchdConfig := &launchd.Config{
		Label:      m.launchConfig.Label,
		ScriptPath: m.launchConfig.ScriptPath,
		Hour:       hour,
		Minute:     minute,
		RunAtLoad:  m.launchConfig.RunAtLoad,
	}
	
	if err := m.launchdMgr.GeneratePlist(launchdConfig); err != nil {
		m.err = fmt.Errorf("failed to generate plist: %w", err)
		return m, nil
	}

	// Load the agent
	if err := m.launchdMgr.Load(); err != nil {
		m.err = fmt.Errorf("failed to load LaunchAgent: %w", err)
		return m, nil
	}

	m.complete = true
	return m, nil
}
