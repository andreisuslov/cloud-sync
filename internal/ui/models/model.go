package models

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
)

// AppState represents the current state of the application
type AppState int

const (
	StateMainMenu AppState = iota
	StateInstallation
	StateConfiguration
	StateBackupRunning
	StateLogViewer
	StateLaunchdManager
	StateMaintenance
	StateHelp
	StateExiting
)

// String returns the string representation of AppState
func (s AppState) String() string {
	switch s {
	case StateMainMenu:
		return "Main Menu"
	case StateInstallation:
		return "Installation"
	case StateConfiguration:
		return "Configuration"
	case StateBackupRunning:
		return "Backup Running"
	case StateLogViewer:
		return "Log Viewer"
	case StateLaunchdManager:
		return "LaunchAgent Manager"
	case StateMaintenance:
		return "Maintenance"
	case StateHelp:
		return "Help"
	case StateExiting:
		return "Exiting"
	default:
		return "Unknown"
	}
}

// Model represents the main application model
type Model struct {
	State         AppState
	List          list.Model
	Viewport      viewport.Model
	Spinner       spinner.Model
	Width         int
	Height        int
	Err           error
	Message       string
	ShowMessage   bool
	Quitting      bool
	
	// Sub-models for different views
	InstallationState *InstallationModel
	ConfigState       *ConfigurationModel
	BackupState       *BackupModel
	LogState          *LogViewerModel
	LaunchdState      *LaunchdModel
	MaintenanceState  *MaintenanceModel
}

// InstallationModel represents the installation wizard state
type InstallationModel struct {
	Step            int
	HomebrewStatus  string
	RcloneStatus    string
	Installing      bool
	Complete        bool
}

// ConfigurationModel represents the configuration wizard state
type ConfigurationModel struct {
	Step            int
	B2Remote        RemoteConfig
	ScalewayRemote  RemoteConfig
	SourceBucket    string
	DestBucket      string
	Complete        bool
}

// RemoteConfig holds configuration for a remote
type RemoteConfig struct {
	Name     string
	Type     string
	KeyID    string
	Key      string
	Region   string
	Provider string
}

// BackupModel represents the backup operation state
type BackupModel struct {
	Running       bool
	Progress      float64
	CurrentFile   string
	TotalFiles    int
	Transferred   int
	Speed         string
	Elapsed       string
	ETA           string
	Complete      bool
	Success       bool
	ErrorMessage  string
}

// LogViewerModel represents the log viewer state
type LogViewerModel struct {
	Filter      string
	SearchTerm  string
	ScrollPos   int
	TotalLines  int
}

// LaunchdModel represents the LaunchAgent manager state
type LaunchdModel struct {
	Loaded      bool
	Running     bool
	PID         int
	LastRun     string
	NextRun     string
	LastExit    int
}

// MaintenanceModel represents maintenance operations state
type MaintenanceModel struct {
	Operation   string
	Confirming  bool
	Complete    bool
	Message     string
}