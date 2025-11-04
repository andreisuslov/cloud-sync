package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andreisuslov/cloud-sync/internal/config"
	"github.com/andreisuslov/cloud-sync/internal/installer"
	"github.com/andreisuslov/cloud-sync/internal/launchd"
	"github.com/andreisuslov/cloud-sync/internal/syncconfig"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
)

// InstallMenuItem represents a menu item in the installation wizard
type InstallMenuItem struct {
	title       string
	description string
}

func (i InstallMenuItem) FilterValue() string { return i.title }
func (i InstallMenuItem) Title() string       { return i.title }
func (i InstallMenuItem) Description() string { return i.description }

// InstallationStep represents a step in the installation wizard
type InstallationStep int

const (
	StepMainInstallMenu InstallationStep = iota
	StepInstallTools
	StepCheckHomebrew
	StepInstallHomebrew
	StepCheckRclone
	StepInstallRclone
	StepCreateDirectories
	StepSetupLocation
	StepLocationTypeSelection
	StepConfigureLocalLocation
	StepConfigureRemoteLocation
	StepConfigureB2Remote
	StepConfigureScalewayRemote
	StepConfigureS3Remote
	StepConfigureSyncPairs
	StepConfigureLaunchAgent
	StepGenerateScripts
	InstallStepComplete
	StepPostInstallMenu
	StepViewConfigs
	StepViewLocations
	StepManageSyncPairs
)

// InstallationModel represents the installation wizard state
type InstallationModel struct {
	installer        *installer.Installer
	configManager    *config.Manager
	launchdMgr       *launchd.Manager
	syncConfigMgr    *syncconfig.Manager
	currentStep      InstallationStep
	spinner          spinner.Model
	mainMenuList     list.Model
	locationTypeList list.Model
	remoteTypeList   list.Model
	homebrewOK       bool
	rcloneOK         bool
	directoriesOK    bool
	b2Configured     bool
	scalewayConfigured bool
	s3Configured     bool
	localConfigured  bool
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
	menuChoice       int
	locationType     string // "local" or "remote"
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
	syncConfigMgr, _ := syncconfig.NewDefaultManager()

	// Create main menu items
	mainMenuItems := []list.Item{
		InstallMenuItem{
			title:       "Install and set up required tools",
			description: "Install Homebrew, rclone, and create necessary directories",
		},
		InstallMenuItem{
			title:       "Set up a new location",
			description: "Configure a local folder or remote storage location",
		},
		InstallMenuItem{
			title:       "View existing locations",
			description: "See all configured remote and local sync locations",
		},
	}

	// Create location type menu items
	locationTypeItems := []list.Item{
		InstallMenuItem{
			title:       "Local folder",
			description: "Configure a local directory for syncing",
		},
		InstallMenuItem{
			title:       "Remote storage (B2, S3, etc.)",
			description: "Configure cloud storage provider",
		},
	}

	// Create remote type menu items - comprehensive list of rclone providers
	remoteTypeItems := []list.Item{
		InstallMenuItem{
			title:       "Backblaze B2",
			description: "High-performance cloud storage",
		},
		InstallMenuItem{
			title:       "Amazon S3",
			description: "AWS Simple Storage Service",
		},
		InstallMenuItem{
			title:       "Google Cloud Storage",
			description: "GCS object storage",
		},
		InstallMenuItem{
			title:       "Microsoft Azure Blob Storage",
			description: "Azure cloud storage",
		},
		InstallMenuItem{
			title:       "Dropbox",
			description: "Cloud file storage and sharing",
		},
		InstallMenuItem{
			title:       "Google Drive",
			description: "Google's cloud storage service",
		},
		InstallMenuItem{
			title:       "OneDrive",
			description: "Microsoft cloud storage",
		},
		InstallMenuItem{
			title:       "Scaleway Object Storage",
			description: "European cloud storage provider",
		},
		InstallMenuItem{
			title:       "DigitalOcean Spaces",
			description: "S3-compatible object storage",
		},
		InstallMenuItem{
			title:       "Wasabi",
			description: "Hot cloud storage",
		},
		InstallMenuItem{
			title:       "SFTP",
			description: "SSH File Transfer Protocol",
		},
		InstallMenuItem{
			title:       "FTP",
			description: "File Transfer Protocol",
		},
		InstallMenuItem{
			title:       "WebDAV",
			description: "Web Distributed Authoring and Versioning",
		},
		InstallMenuItem{
			title:       "Other / Custom",
			description: "Configure any other rclone-supported provider",
		},
	}

	// Create lists
	mainMenuList := list.New(mainMenuItems, list.NewDefaultDelegate(), 0, 0)
	mainMenuList.Title = "Installation Menu"
	mainMenuList.Styles.Title = styles.TitleStyle
	mainMenuList.SetShowHelp(false)

	locationTypeList := list.New(locationTypeItems, list.NewDefaultDelegate(), 0, 0)
	locationTypeList.Title = "Set Up a New Location"
	locationTypeList.Styles.Title = styles.TitleStyle
	locationTypeList.SetShowHelp(false)

	remoteTypeList := list.New(remoteTypeItems, list.NewDefaultDelegate(), 0, 0)
	remoteTypeList.Title = "Configure Remote Storage"
	remoteTypeList.Styles.Title = styles.TitleStyle
	remoteTypeList.SetShowHelp(false)

	return InstallationModel{
		installer:        installer.NewInstaller(),
		configManager:    configMgr,
		launchdMgr:       launchdMgr,
		syncConfigMgr:    syncConfigMgr,
		currentStep:      StepMainInstallMenu,
		spinner:          s,
		mainMenuList:     mainMenuList,
		locationTypeList: locationTypeList,
		remoteTypeList:   remoteTypeList,
		homeDir:          homeDir,
		binDir:           filepath.Join(homeDir, "bin"),
		logDir:           filepath.Join(homeDir, "logs"),
		menuChoice:       0,
	}
}

// Init initializes the installation wizard
func (m InstallationModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the installation wizard
func (m InstallationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update list sizes
		listHeight := msg.Height - 8
		if listHeight < 5 {
			listHeight = 5
		}
		m.mainMenuList.SetSize(msg.Width-4, listHeight)
		m.locationTypeList.SetSize(msg.Width-4, listHeight)
		m.remoteTypeList.SetSize(msg.Width-4, listHeight)
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
				case StepManageSyncPairs:
					if msg.String() == "q" || msg.String() == "esc" {
						m.activeSubView = nil
						m.currentStep = StepPostInstallMenu
						return m, nil
					}
				}
			}
			return m, cmd
		}
		
		// Handle main installation menu navigation
		if m.currentStep == StepMainInstallMenu {
			switch msg.String() {
			case "enter":
				return m.handleMainMenuSelection()
			case "q", "esc":
				return m, tea.Quit
			default:
				// Let the list handle navigation
				var cmd tea.Cmd
				m.mainMenuList, cmd = m.mainMenuList.Update(msg)
				return m, cmd
			}
		} else if m.currentStep == StepLocationTypeSelection {
			switch msg.String() {
			case "enter":
				return m.handleLocationTypeSelection()
			case "q", "esc":
				m.currentStep = StepMainInstallMenu
				return m, nil
			default:
				// Let the list handle navigation
				var cmd tea.Cmd
				m.locationTypeList, cmd = m.locationTypeList.Update(msg)
				return m, cmd
			}
		} else if m.currentStep == StepConfigureRemoteLocation {
			switch msg.String() {
			case "enter":
				return m.handleRemoteTypeSelection()
			case "q", "esc":
				m.currentStep = StepLocationTypeSelection
				return m, nil
			default:
				// Let the list handle navigation
				var cmd tea.Cmd
				m.remoteTypeList, cmd = m.remoteTypeList.Update(msg)
				return m, cmd
			}
		} else if m.currentStep == StepPostInstallMenu {
			switch msg.String() {
			case "1":
				m.currentStep = StepViewConfigs
				return m, nil
			case "2":
				m.currentStep = StepManageSyncPairs
				syncPairsModel := NewSyncPairsModel(m.syncConfigMgr)
				m.activeSubView = syncPairsModel
				return m, syncPairsModel.Init()
			case "q", "esc":
				return m, tea.Quit
			}
		} else if m.currentStep == StepViewConfigs || m.currentStep == StepViewLocations {
			if msg.String() == "q" || msg.String() == "esc" {
				m.currentStep = StepMainInstallMenu
				return m, nil
			}
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

	// For list-based menus, render them directly without header/box
	if m.currentStep == StepMainInstallMenu || m.currentStep == StepLocationTypeSelection || m.currentStep == StepConfigureRemoteLocation {
		b.WriteString("\n")
		b.WriteString(m.renderCurrentStep())
		b.WriteString("\n\n")
		
		// Render footer based on current step
		switch m.currentStep {
		case StepMainInstallMenu:
			b.WriteString(helper.RenderFooter("↑/↓: Navigate • Enter: Select • q: Exit"))
		case StepLocationTypeSelection:
			b.WriteString(helper.RenderFooter("↑/↓: Navigate • Enter: Select • q: Back"))
		case StepConfigureRemoteLocation:
			b.WriteString(helper.RenderFooter("↑/↓: Navigate • Enter: Select • q: Back"))
		}
		
		return b.String()
	}

	b.WriteString(helper.RenderHeader("Installation Wizard", "Setting up required tools"))

	// Show current progress (but not for post-install menu)
	if m.currentStep < StepPostInstallMenu {
		b.WriteString(m.renderProgress())
		b.WriteString("\n\n")
	}

	// Show current step
	b.WriteString(m.renderCurrentStep())
	b.WriteString("\n")

	if m.error != nil {
		b.WriteString("\n")
		b.WriteString(styles.RenderError(fmt.Sprintf("Error: %v", m.error)))
		b.WriteString("\n")
	}

	// Render footer based on current step
	switch m.currentStep {
	case StepViewLocations:
		b.WriteString(helper.RenderFooter("q: Back to menu"))
	case StepPostInstallMenu:
		b.WriteString(helper.RenderFooter("Select option (1-2) • q: Exit"))
	case StepViewConfigs:
		b.WriteString(helper.RenderFooter("q: Back to menu"))
	default:
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
	case StepMainInstallMenu:
		content = m.renderMainInstallMenu()
	case StepLocationTypeSelection:
		content = m.renderLocationTypeSelection()
	case StepConfigureRemoteLocation:
		content = m.renderRemoteLocationSelection()
	case StepViewLocations:
		content = m.renderLocationsView()
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
	case StepPostInstallMenu:
		content = m.renderPostInstallMenu()
	case StepViewConfigs:
		content = m.renderConfigsView()
	}

	return box.Render(content)
}

// renderPostInstallMenu renders the post-install menu
func (m InstallationModel) renderPostInstallMenu() string {
	var b strings.Builder
	b.WriteString(styles.RenderSuccess("✓ Installation Complete!\n\n"))
	b.WriteString("What would you like to do?\n\n")
	b.WriteString("1. View current rclone remotes and sync configurations\n")
	b.WriteString("2. Add/manage sync pairs\n")
	b.WriteString("\nPress 1 or 2 to select, q to exit")
	return b.String()
}

// renderMainInstallMenu renders the main installation menu
func (m InstallationModel) renderMainInstallMenu() string {
	return m.mainMenuList.View()
}

// renderLocationTypeSelection renders the location type selection menu
func (m InstallationModel) renderLocationTypeSelection() string {
	return m.locationTypeList.View()
}

// renderRemoteLocationSelection renders the remote location type selection menu
func (m InstallationModel) renderRemoteLocationSelection() string {
	return m.remoteTypeList.View()
}

// renderLocationsView renders the existing locations view
func (m InstallationModel) renderLocationsView() string {
	var b strings.Builder
	
	b.WriteString(styles.RenderInfo("Existing Locations\n\n"))
	
	// Show rclone remotes
	b.WriteString("Remote Locations:\n")
	appConfig, err := m.configManager.Load()
	if err != nil || len(appConfig.Remotes) == 0 {
		b.WriteString("  No remote locations configured\n")
	} else {
		for _, remote := range appConfig.Remotes {
			b.WriteString(fmt.Sprintf("  • %s (%s)\n", remote.Name, remote.Provider))
		}
	}
	
	b.WriteString("\nLocal Sync Folders:\n")
	syncPairs, err := m.syncConfigMgr.ListSyncPairs()
	if err != nil || len(syncPairs) == 0 {
		b.WriteString("  No local folders configured\n")
	} else {
		// Show unique local paths
		localPaths := make(map[string]bool)
		for _, pair := range syncPairs {
			if !localPaths[pair.LocalPath] {
				localPaths[pair.LocalPath] = true
				status := "✓"
				if !pair.Enabled {
					status = "✗"
				}
				b.WriteString(fmt.Sprintf("  [%s] %s\n", status, pair.LocalPath))
			}
		}
	}
	
	b.WriteString("\nPress q or Esc to go back")
	
	return b.String()
}

// renderConfigsView renders the current configurations
func (m InstallationModel) renderConfigsView() string {
	var b strings.Builder
	
	b.WriteString(styles.RenderInfo("Current Configuration\n\n"))
	
	// Show rclone remotes
	b.WriteString("Rclone Remotes:\n")
	appConfig, err := m.configManager.Load()
	if err != nil || len(appConfig.Remotes) == 0 {
		b.WriteString("  No remotes configured\n")
	} else {
		for _, remote := range appConfig.Remotes {
			b.WriteString(fmt.Sprintf("  • %s (%s)\n", remote.Name, remote.Provider))
		}
	}
	
	b.WriteString("\nSync Pairs:\n")
	syncPairs, err := m.syncConfigMgr.ListSyncPairs()
	if err != nil || len(syncPairs) == 0 {
		b.WriteString("  No sync pairs configured\n")
	} else {
		for _, pair := range syncPairs {
			status := "✓"
			if !pair.Enabled {
				status = "✗"
			}
			b.WriteString(fmt.Sprintf("  [%s] %s: %s → %s:%s (%s)\n", 
				status, pair.Name, pair.LocalPath, pair.RemoteName, pair.RemotePath, pair.Direction))
		}
	}
	
	b.WriteString("\n\nCommands:\n")
	b.WriteString("  View rclone config: rclone config\n")
	b.WriteString("  List remotes: rclone listremotes\n")
	b.WriteString("  Test remote: rclone lsd <remote>:\n")
	b.WriteString("\nPress q or Esc to go back")
	
	return b.String()
}

// handleMainMenuSelection handles main menu item selection
func (m InstallationModel) handleMainMenuSelection() (tea.Model, tea.Cmd) {
	selected := m.mainMenuList.SelectedItem()
	if selected == nil {
		return m, nil
	}

	menuItem := selected.(InstallMenuItem)
	title := menuItem.Title()

	switch title {
	case "Install and set up required tools":
		m.currentStep = StepInstallTools
		m.installing = true
		return m, m.checkHomebrew()
	case "Set up a new location":
		m.currentStep = StepLocationTypeSelection
		return m, nil
	case "View existing locations":
		m.currentStep = StepViewLocations
		return m, nil
	}

	return m, nil
}

// handleLocationTypeSelection handles location type selection
func (m InstallationModel) handleLocationTypeSelection() (tea.Model, tea.Cmd) {
	selected := m.locationTypeList.SelectedItem()
	if selected == nil {
		return m, nil
	}

	menuItem := selected.(InstallMenuItem)
	title := menuItem.Title()

	switch title {
	case "Local folder":
		m.locationType = "local"
		m.currentStep = StepConfigureLocalLocation
		// TODO: Initialize local location config view
		return m, nil
	case "Remote storage (B2, S3, etc.)":
		m.locationType = "remote"
		m.currentStep = StepConfigureRemoteLocation
		return m, nil
	}

	return m, nil
}

// handleRemoteTypeSelection handles remote type selection
func (m InstallationModel) handleRemoteTypeSelection() (tea.Model, tea.Cmd) {
	selected := m.remoteTypeList.SelectedItem()
	if selected == nil {
		return m, nil
	}

	menuItem := selected.(InstallMenuItem)
	providerName := menuItem.Title()

	// Map provider display names to rclone types and configuration (for future use)
	_ = map[string]struct {
		rcloneType string
		step       InstallationStep
	}{
		"Backblaze B2":                 {rcloneType: "b2", step: StepConfigureB2Remote},
		"Amazon S3":                    {rcloneType: "s3", step: StepConfigureS3Remote},
		"Scaleway Object Storage":      {rcloneType: "s3", step: StepConfigureScalewayRemote},
		"Google Cloud Storage":         {rcloneType: "google cloud storage", step: StepConfigureRemoteLocation},
		"Microsoft Azure Blob Storage": {rcloneType: "azureblob", step: StepConfigureRemoteLocation},
		"Dropbox":                      {rcloneType: "dropbox", step: StepConfigureRemoteLocation},
		"Google Drive":                 {rcloneType: "drive", step: StepConfigureRemoteLocation},
		"OneDrive":                     {rcloneType: "onedrive", step: StepConfigureRemoteLocation},
		"DigitalOcean Spaces":          {rcloneType: "s3", step: StepConfigureRemoteLocation},
		"Wasabi":                       {rcloneType: "s3", step: StepConfigureRemoteLocation},
		"SFTP":                         {rcloneType: "sftp", step: StepConfigureRemoteLocation},
		"FTP":                          {rcloneType: "ftp", step: StepConfigureRemoteLocation},
		"WebDAV":                       {rcloneType: "webdav", step: StepConfigureRemoteLocation},
		"Other / Custom":               {rcloneType: "custom", step: StepConfigureRemoteLocation},
	}

	// For now, handle the providers we have specific configuration for
	switch providerName {
	case "Backblaze B2":
		m.currentStep = StepConfigureB2Remote
		return m, m.initRemoteConfig(providerName)
	case "Scaleway Object Storage":
		m.currentStep = StepConfigureScalewayRemote
		return m, m.initRemoteConfig(providerName)
	default:
		// For all other providers, initialize a generic remote config
		m.currentStep = StepConfigureRemoteLocation
		return m, m.initRemoteConfig(providerName)
	}
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
	m.currentStep = StepPostInstallMenu
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

// initRemoteConfig initializes the remote configuration sub-view for any provider
func (m InstallationModel) initRemoteConfig(providerName string) tea.Cmd {
	return func() tea.Msg {
		m.activeSubView = NewRemoteConfigModelWithProvider(m.configManager, providerName)
		return m.activeSubView.Init()
	}
}

// initB2Config initializes the B2 configuration sub-view (legacy)
func (m InstallationModel) initB2Config() tea.Cmd {
	return m.initRemoteConfig("Backblaze B2")
}

// initScalewayConfig initializes the Scaleway configuration sub-view (legacy)
func (m InstallationModel) initScalewayConfig() tea.Cmd {
	return m.initRemoteConfig("Scaleway Object Storage")
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
