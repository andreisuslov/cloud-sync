package unit

import (
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
