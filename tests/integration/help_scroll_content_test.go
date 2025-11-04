package integration

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/andreisuslov/cloud-sync/internal/ui"
)

// TestHelpMenuScrollContentVisibility tests the specific issue where
// the help menu title and content were being clipped when scrolling
func TestHelpMenuScrollContentVisibility(t *testing.T) {
	assert := assert.New(t)

	// Initialize model
	m := ui.NewModel()
	mAny, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 20})
	m = mAny.(ui.Model)

	// Navigate to Help menu (item 6, which is index 5)
	for i := 0; i < 5; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
	}
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mAny.(ui.Model)

	// Test Case 1: Initial view at 0% scroll should show title
	view := stripANSI(m.View())
	assert.Contains(view, "Keyboard Shortcuts & Help", "Initial view should contain title")
	assert.Contains(view, "========================", "Initial view should contain separator")
	assert.Contains(view, "Global Shortcuts:", "Initial view should contain Global Shortcuts section")
	assert.Contains(view, "Scroll: 0%", "Initial scroll should be at 0%")

	// Test Case 2: Scroll down 5 times
	for i := 0; i < 5; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
	}
	view = stripANSI(m.View())
	
	// After scrolling down, we should see different content
	// The title might not be visible anymore, but we should see later sections
	assert.Contains(view, "Main Menu:", "After scrolling down, should see Main Menu section")
	
	// Verify scroll percentage increased
	scrollPercent := extractScrollPercent(view)
	assert.Greater(scrollPercent, 0, "Scroll percentage should be greater than 0%")

	// Test Case 3: Scroll up 2 times from current position
	for i := 0; i < 2; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = mAny.(ui.Model)
	}
	view = stripANSI(m.View())
	
	// Scroll percentage should have decreased
	newScrollPercent := extractScrollPercent(view)
	assert.Less(newScrollPercent, scrollPercent, "Scroll percentage should decrease when scrolling up")

	// Test Case 4: Scroll up 1 more time
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = mAny.(ui.Model)
	view = stripANSI(m.View())
	
	scrollPercent = extractScrollPercent(view)
	assert.Less(scrollPercent, newScrollPercent, "Scroll percentage should continue to decrease")

	// Test Case 5: Scroll up 1 more time (should be near top)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = mAny.(ui.Model)
	view = stripANSI(m.View())
	
	finalScrollPercent := extractScrollPercent(view)
	assert.Less(finalScrollPercent, scrollPercent, "Scroll percentage should continue to decrease")
	
	// At low scroll percentages, we should see the title and separator
	// This is the key test - ensuring content isn't clipped at the top
	if finalScrollPercent <= 5 {
		// When we're at 3-5%, we should see at least part of the beginning content
		hasTitle := strings.Contains(view, "Keyboard Shortcuts")
		hasSeparator := strings.Contains(view, "========================")
		hasGlobalShortcuts := strings.Contains(view, "Global Shortcuts:")
		
		// At least one of these should be visible when near the top
		assert.True(hasTitle || hasSeparator || hasGlobalShortcuts,
			"When scroll is at %d%%, at least some top content should be visible", finalScrollPercent)
	}

	// Test Case 6: Continue scrolling up to reach the very top
	// Keep pressing up until we reach 0%
	for i := 0; i < 10; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = mAny.(ui.Model)
	}
	view = stripANSI(m.View())
	
	// At the very top, all initial content must be visible
	assert.Contains(view, "Keyboard Shortcuts & Help", "At top, title must be visible")
	assert.Contains(view, "========================", "At top, separator must be visible")
	assert.Contains(view, "Global Shortcuts:", "At top, Global Shortcuts must be visible")
	assert.Contains(view, "Scroll: 0%", "At top, scroll should be 0%")
}

// TestHelpMenuScrollToBottomAndBack tests scrolling to the bottom and back to top
func TestHelpMenuScrollToBottomAndBack(t *testing.T) {
	assert := assert.New(t)

	// Initialize model
	m := ui.NewModel()
	mAny, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 20})
	m = mAny.(ui.Model)

	// Navigate to Help menu
	for i := 0; i < 5; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
	}
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mAny.(ui.Model)

	// Scroll to bottom by pressing down many times
	for i := 0; i < 50; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
	}
	view := stripANSI(m.View())
	
	// Should show bottom content
	assert.Contains(view, "GitHub:", "At bottom, should see GitHub info")
	scrollPercent := extractScrollPercent(view)
	assert.Equal(100, scrollPercent, "At bottom, scroll should be 100%")

	// Scroll back to top by pressing up many times
	for i := 0; i < 50; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = mAny.(ui.Model)
	}
	view = stripANSI(m.View())
	
	// Should show top content again
	assert.Contains(view, "Keyboard Shortcuts & Help", "After returning to top, title must be visible")
	assert.Contains(view, "========================", "After returning to top, separator must be visible")
	assert.Contains(view, "Scroll: 0%", "After returning to top, scroll should be 0%")
}

// TestHelpMenuPageUpDown tests page up/down navigation
func TestHelpMenuPageUpDown(t *testing.T) {
	assert := assert.New(t)

	// Initialize model
	m := ui.NewModel()
	mAny, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 20})
	m = mAny.(ui.Model)

	// Navigate to Help menu
	for i := 0; i < 5; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
	}
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mAny.(ui.Model)

	// Get initial scroll position
	view := stripANSI(m.View())
	initialScroll := extractScrollPercent(view)

	// Page down
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = mAny.(ui.Model)
	view = stripANSI(m.View())
	afterPageDown := extractScrollPercent(view)
	
	assert.Greater(afterPageDown, initialScroll, "Page down should increase scroll position")

	// Page up
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = mAny.(ui.Model)
	view = stripANSI(m.View())
	afterPageUp := extractScrollPercent(view)
	
	assert.Less(afterPageUp, afterPageDown, "Page up should decrease scroll position")
}

// extractScrollPercent extracts the scroll percentage from the view
// Looks for pattern like "Scroll: 14%"
func extractScrollPercent(view string) int {
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Scroll:") {
			// Extract percentage
			parts := strings.Split(line, "Scroll:")
			if len(parts) > 1 {
				percentStr := strings.TrimSpace(parts[1])
				percentStr = strings.TrimSuffix(percentStr, "%")
				percentStr = strings.Split(percentStr, " ")[0]
				var percent int
				_, err := fmt.Sscanf(percentStr, "%d", &percent)
				if err == nil {
					return percent
				}
			}
		}
	}
	return -1 // Not found
}
