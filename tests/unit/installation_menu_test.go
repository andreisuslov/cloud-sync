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
	
	// Test main menu rendering - should be empty now
	view := model.View()
	if !strings.Contains(view, "Installation Menu") {
		t.Error("Main menu should display 'Installation Menu'")
	}
	// Installation menu is now empty - no items to check
}

func TestLocationTypeSelection(t *testing.T) {
	t.Skip("Installation menu is now empty - test no longer applicable")
}

func TestRemoteLocationSelection(t *testing.T) {
	t.Skip("Installation menu is now empty - test no longer applicable")
}

func TestViewExistingLocations(t *testing.T) {
	t.Skip("Installation menu is now empty - test no longer applicable")
}

func TestBackNavigation(t *testing.T) {
	t.Skip("Installation menu is now empty - test no longer applicable")
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
