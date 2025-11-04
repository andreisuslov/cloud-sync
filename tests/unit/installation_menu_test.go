package unit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andreisuslov/cloud-sync/internal/ui/views"
)

func TestInstallationPostMenu(t *testing.T) {
	// Create installation model
	model := views.NewInstallationModel()
	
	// Simulate window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(views.InstallationModel)
	
	// Test that model initializes correctly
	if model.Init() == nil {
		t.Error("Init should return a command")
	}
}

func TestInstallationMenuNavigation(t *testing.T) {
	model := views.NewInstallationModel()
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(views.InstallationModel)
	
	// Test view rendering doesn't panic
	view := model.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMainInstallationMenu(t *testing.T) {
	model := views.NewInstallationModel()
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	model = updatedModel.(views.InstallationModel)
	
	// Test main menu rendering
	view := model.View()
	if !strings.Contains(view, "Installation Menu") {
		t.Error("Main menu should display 'Installation Menu'")
	}
	if !strings.Contains(view, "Install and set up required tools") {
		t.Error("Main menu should display option 1")
	}
	if !strings.Contains(view, "Set up a new location") {
		t.Error("Main menu should display option 2")
	}
	if !strings.Contains(view, "View existing locations") {
		t.Error("Main menu should display option 3")
	}
}

func TestLocationTypeSelection(t *testing.T) {
	model := views.NewInstallationModel()
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(views.InstallationModel)
	
	// Navigate down to "Set up a new location" and select it
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(views.InstallationModel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(views.InstallationModel)
	
	// Test location type selection rendering
	view := model.View()
	if !strings.Contains(view, "Set Up a New Location") {
		t.Error("Should display 'Set Up a New Location' heading")
	}
	if !strings.Contains(view, "Local folder") {
		t.Error("Should display 'Local folder' option")
	}
	if !strings.Contains(view, "Remote storage") {
		t.Error("Should display 'Remote storage' option")
	}
}

func TestRemoteLocationSelection(t *testing.T) {
	model := views.NewInstallationModel()
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(views.InstallationModel)
	
	// Navigate to "Set up a new location" and select it
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(views.InstallationModel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(views.InstallationModel)
	
	// Navigate to "Remote storage" and select it
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(views.InstallationModel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(views.InstallationModel)
	
	// Test remote location selection rendering
	view := model.View()
	if !strings.Contains(view, "Configure Remote Storage") {
		t.Error("Should display 'Configure Remote Storage' heading")
	}
	if !strings.Contains(view, "Backblaze B2") {
		t.Error("Should display 'Backblaze B2' option")
	}
	if !strings.Contains(view, "Amazon S3") {
		t.Error("Should display 'Amazon S3' option")
	}
}

func TestViewExistingLocations(t *testing.T) {
	model := views.NewInstallationModel()
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(views.InstallationModel)
	
	// Navigate to "View existing locations" and select it
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(views.InstallationModel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(views.InstallationModel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(views.InstallationModel)
	
	// Test existing locations view rendering
	view := model.View()
	if !strings.Contains(view, "Existing Locations") {
		t.Error("Should display 'Existing Locations' heading")
	}
	if !strings.Contains(view, "Remote Locations:") {
		t.Error("Should display 'Remote Locations:' section")
	}
	if !strings.Contains(view, "Local Sync Folders:") {
		t.Error("Should display 'Local Sync Folders:' section")
	}
}

func TestBackNavigation(t *testing.T) {
	model := views.NewInstallationModel()
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(views.InstallationModel)
	
	// Navigate to "Set up a new location" and select it
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(views.InstallationModel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(views.InstallationModel)
	
	// Navigate back with 'q'
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updatedModel.(views.InstallationModel)
	
	// Should be back at main menu
	view := model.View()
	if !strings.Contains(view, "Installation Menu") {
		t.Error("Should return to main menu after pressing 'q'")
	}
}

func TestConfigsViewRendering(t *testing.T) {
	model := views.NewInstallationModel()
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(views.InstallationModel)
	
	// Test that view renders without errors
	view := model.View()
	if len(view) == 0 {
		t.Error("View should render content")
	}
}
