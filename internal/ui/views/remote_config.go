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

// RemoteConfigStep represents a step in the remote configuration wizard
type RemoteConfigStep int

const (
	RemoteStepSelectType RemoteConfigStep = iota
	RemoteStepB2Config
	RemoteStepScalewayConfig
	RemoteStepComplete
)

// RemoteConfigModel represents the remote configuration wizard
type RemoteConfigModel struct {
	currentStep   RemoteConfigStep
	remoteType    string // "b2" or "s3"
	inputs        []textinput.Model
	focusIndex    int
	width         int
	height        int
	err           error
	complete      bool
	configManager *config.Manager
	remoteConfig  config.RemoteConfig
}

// NewRemoteConfigModel creates a new remote configuration model
func NewRemoteConfigModel(configManager *config.Manager) RemoteConfigModel {
	return RemoteConfigModel{
		currentStep:   RemoteStepSelectType,
		configManager: configManager,
		inputs:        make([]textinput.Model, 0),
	}
}

// Init initializes the remote configuration wizard
func (m RemoteConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the remote configuration wizard
func (m RemoteConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.currentStep == RemoteStepSelectType {
				return m, nil
			}
			// Go back to previous step
			if m.currentStep > RemoteStepSelectType {
				m.currentStep--
			}
			return m, nil

		case "tab", "shift+tab", "up", "down":
			if len(m.inputs) > 0 {
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
			}

		case "enter":
			return m.handleEnter()

		case "1":
			if m.currentStep == RemoteStepSelectType {
				m.remoteType = "b2"
				m.currentStep = RemoteStepB2Config
				return m, m.initB2Inputs()
			}

		case "2":
			if m.currentStep == RemoteStepSelectType {
				m.remoteType = "s3"
				m.currentStep = RemoteStepScalewayConfig
				return m, m.initScalewayInputs()
			}
		}
	}

	// Handle character input for text fields
	if len(m.inputs) > 0 && m.focusIndex < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the remote configuration wizard
func (m RemoteConfigModel) View() string {
	helper := NewViewHelper(m.width, m.height)
	var b strings.Builder

	b.WriteString(helper.RenderHeader("Configure Rclone Remote", "Set up cloud storage credentials"))

	switch m.currentStep {
	case RemoteStepSelectType:
		b.WriteString(m.renderSelectType())
	case RemoteStepB2Config:
		b.WriteString(m.renderB2Config())
	case RemoteStepScalewayConfig:
		b.WriteString(m.renderScalewayConfig())
	case RemoteStepComplete:
		b.WriteString(m.renderComplete())
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(styles.RenderError(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	if m.currentStep == RemoteStepSelectType {
		b.WriteString(helper.RenderFooter("1: Backblaze B2 • 2: Scaleway • q: Back"))
	} else if m.complete {
		b.WriteString(helper.RenderFooter("Press Enter to continue • q: Back to menu"))
	} else {
		b.WriteString(helper.RenderFooter("Tab: Next field • Enter: Save • q: Back"))
	}

	return b.String()
}

// renderSelectType renders the remote type selection
func (m RemoteConfigModel) renderSelectType() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(2, 4).
		Width(60)

	content := "Select Remote Type:\n\n"
	content += "1. Backblaze B2\n"
	content += "   - Source storage for your data\n\n"
	content += "2. Scaleway Object Storage\n"
	content += "   - Destination for backups\n"

	return box.Render(content)
}

// renderB2Config renders the B2 configuration form
func (m RemoteConfigModel) renderB2Config() string {
	var b strings.Builder
	
	b.WriteString(styles.RenderInfo("Backblaze B2 Configuration"))
	b.WriteString("\n\n")
	
	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n\n")
		}
	}
	
	b.WriteString("\n\n")
	b.WriteString(styles.RenderMuted("Get your credentials from: https://www.backblaze.com/b2/cloud-storage.html"))
	
	return b.String()
}

// renderScalewayConfig renders the Scaleway configuration form
func (m RemoteConfigModel) renderScalewayConfig() string {
	var b strings.Builder
	
	b.WriteString(styles.RenderInfo("Scaleway Object Storage Configuration"))
	b.WriteString("\n\n")
	
	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n\n")
		}
	}
	
	b.WriteString("\n\n")
	b.WriteString(styles.RenderMuted("Get your credentials from: https://console.scaleway.com/"))
	
	return b.String()
}

// renderComplete renders the completion message
func (m RemoteConfigModel) renderComplete() string {
	return styles.RenderSuccess(fmt.Sprintf("✓ Remote '%s' configured successfully!\n\nConfiguration saved.", m.remoteConfig.Name))
}

// initB2Inputs initializes input fields for B2 configuration
func (m RemoteConfigModel) initB2Inputs() tea.Cmd {
	inputs := make([]textinput.Model, 3)

	// Remote name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "b2"
	inputs[0].Focus()
	inputs[0].PromptStyle = styles.FocusedStyle
	inputs[0].TextStyle = styles.FocusedStyle
	inputs[0].CharLimit = 32
	inputs[0].Width = 50
	inputs[0].Prompt = "Remote Name: "

	// Account ID
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Your B2 Account ID"
	inputs[1].CharLimit = 100
	inputs[1].Width = 50
	inputs[1].Prompt = "Account ID: "

	// Application Key
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Your B2 Application Key"
	inputs[2].CharLimit = 100
	inputs[2].Width = 50
	inputs[2].Prompt = "App Key: "
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '•'

	m.inputs = inputs
	m.focusIndex = 0

	return inputs[0].Focus()
}

// initScalewayInputs initializes input fields for Scaleway configuration
func (m RemoteConfigModel) initScalewayInputs() tea.Cmd {
	inputs := make([]textinput.Model, 5)

	// Remote name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "sw"
	inputs[0].Focus()
	inputs[0].PromptStyle = styles.FocusedStyle
	inputs[0].TextStyle = styles.FocusedStyle
	inputs[0].CharLimit = 32
	inputs[0].Width = 50
	inputs[0].Prompt = "Remote Name: "

	// Access Key ID
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Your Scaleway Access Key"
	inputs[1].CharLimit = 100
	inputs[1].Width = 50
	inputs[1].Prompt = "Access Key: "

	// Secret Access Key
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Your Scaleway Secret Key"
	inputs[2].CharLimit = 100
	inputs[2].Width = 50
	inputs[2].Prompt = "Secret Key: "
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '•'

	// Region
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "nl-ams"
	inputs[3].CharLimit = 20
	inputs[3].Width = 50
	inputs[3].Prompt = "Region: "
	inputs[3].SetValue("nl-ams")

	// Endpoint
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "s3.nl-ams.scw.cloud"
	inputs[4].CharLimit = 100
	inputs[4].Width = 50
	inputs[4].Prompt = "Endpoint: "
	inputs[4].SetValue("s3.nl-ams.scw.cloud")

	m.inputs = inputs
	m.focusIndex = 0

	return inputs[0].Focus()
}

// handleEnter handles the Enter key press
func (m RemoteConfigModel) handleEnter() (tea.Model, tea.Cmd) {
	if m.currentStep == RemoteStepComplete {
		return m, nil
	}

	// Validate and save configuration
	if m.currentStep == RemoteStepB2Config {
		if len(m.inputs) < 3 {
			m.err = fmt.Errorf("invalid input configuration")
			return m, nil
		}

		name := strings.TrimSpace(m.inputs[0].Value())
		accountID := strings.TrimSpace(m.inputs[1].Value())
		appKey := strings.TrimSpace(m.inputs[2].Value())

		if name == "" || accountID == "" || appKey == "" {
			m.err = fmt.Errorf("all fields are required")
			return m, nil
		}

		m.remoteConfig = config.RemoteConfig{
			Name:           name,
			Type:           "b2",
			Provider:       "Backblaze",
			AccountID:      accountID,
			ApplicationKey: appKey,
		}

		if err := m.configManager.AddRemote(m.remoteConfig); err != nil {
			m.err = err
			return m, nil
		}

		// Generate rclone config
		if err := m.configManager.GenerateRcloneConfig(); err != nil {
			m.err = err
			return m, nil
		}

		m.currentStep = RemoteStepComplete
		m.complete = true
		return m, nil

	} else if m.currentStep == RemoteStepScalewayConfig {
		if len(m.inputs) < 5 {
			m.err = fmt.Errorf("invalid input configuration")
			return m, nil
		}

		name := strings.TrimSpace(m.inputs[0].Value())
		accessKey := strings.TrimSpace(m.inputs[1].Value())
		secretKey := strings.TrimSpace(m.inputs[2].Value())
		region := strings.TrimSpace(m.inputs[3].Value())
		endpoint := strings.TrimSpace(m.inputs[4].Value())

		if name == "" || accessKey == "" || secretKey == "" {
			m.err = fmt.Errorf("name, access key, and secret key are required")
			return m, nil
		}

		m.remoteConfig = config.RemoteConfig{
			Name:           name,
			Type:           "s3",
			Provider:       "Scaleway",
			AccountID:      accessKey,
			ApplicationKey: secretKey,
			Region:         region,
			Endpoint:       endpoint,
		}

		if err := m.configManager.AddRemote(m.remoteConfig); err != nil {
			m.err = err
			return m, nil
		}

		// Generate rclone config
		if err := m.configManager.GenerateRcloneConfig(); err != nil {
			m.err = err
			return m, nil
		}

		m.currentStep = RemoteStepComplete
		m.complete = true
		return m, nil
	}

	return m, nil
}
