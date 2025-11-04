package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/andreisuslov/cloud-sync/internal/ui"
)

func stripANSI(s string) string {
	ansi := regexp.MustCompile("\\x1b\\[[0-9;]*[a-zA-Z]")
	return ansi.ReplaceAllString(s, "")
}

func writeSnapshot(t *testing.T, dir string, step int, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, fmt.Sprintf("%02d_%s.txt", step, name))
	req := require.New(t)
	req.NoError(os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestMainMenuScrollingSnapshots(t *testing.T) {
	req := require.New(t)

	m := ui.NewModel()

	// Set a small window to force list to scroll
	mAny, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 10})
	m = mAny.(ui.Model)

	tmp := t.TempDir()

	// Initial render
	view := stripANSI(m.View())
	writeSnapshot(t, tmp, 0, "initial", view)

	// Press Down 8 times, capturing after each
	downs := 8
	indices := make([]int, 0, downs+3)
	indices = append(indices, m.List.Index())
	for i := 1; i <= downs; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
		view = stripANSI(m.View())
		writeSnapshot(t, tmp, i, fmt.Sprintf("down_%02d", i), view)
		indices = append(indices, m.List.Index())
	}

	// Ensure index increased monotonically until it can't
	for i := 1; i < len(indices); i++ {
		req.GreaterOrEqual(indices[i], indices[i-1])
	}

	// Now press Up 3 times
	ups := 3
	for j := 1; j <= ups; j++ {
		step := downs + j
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = mAny.(ui.Model)
		view = stripANSI(m.View())
		writeSnapshot(t, tmp, step, fmt.Sprintf("up_%02d", j), view)
		indices = append(indices, m.List.Index())
	}

	// Verify the last three indices moved up or stayed within bounds
	last := len(indices)
	req.LessOrEqual(indices[last-1], indices[last-2])
	req.LessOrEqual(indices[last-2], indices[last-3])

	// Basic difference checks between some snapshots to ensure UI updates
	// Compare a middle-down step with initial
	midDown := filepath.Join(tmp, fmt.Sprintf("%02d_down_%02d.txt", 4, 4))
	b1, err := os.ReadFile(midDown)
	req.NoError(err)
	b0, err := os.ReadFile(filepath.Join(tmp, "00_initial.txt"))
	req.NoError(err)
	req.NotEqual(string(b0), string(b1))

	// Compare last up with previous down
	lastUp := filepath.Join(tmp, fmt.Sprintf("%02d_up_%02d.txt", downs+ups, ups))
	prevDown := filepath.Join(tmp, fmt.Sprintf("%02d_down_%02d.txt", downs, downs))
	bU, err := os.ReadFile(lastUp)
	req.NoError(err)
	bD, err := os.ReadFile(prevDown)
	req.NoError(err)
	req.NotEqual(string(bD), string(bU))
}

func TestHelpViewportScrollingSnapshots(t *testing.T) {
	req := require.New(t)

	m := ui.NewModel()
	mAny, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 12})
	m = mAny.(ui.Model)

	// Enter Help (menu item 6)
	for i := 0; i < 5; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
	}
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mAny.(ui.Model)

	tmp := t.TempDir()

	view := stripANSI(m.View())
	writeSnapshot(t, tmp, 0, "help_initial", view)

	// Scroll down multiple times
	for i := 1; i <= 6; i++ {
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = mAny.(ui.Model)
		view = stripANSI(m.View())
		writeSnapshot(t, tmp, i, fmt.Sprintf("help_down_%02d", i), view)
	}

	// Scroll up 3 times
	for j := 1; j <= 3; j++ {
		step := 6 + j
		mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = mAny.(ui.Model)
		view = stripANSI(m.View())
		writeSnapshot(t, tmp, step, fmt.Sprintf("help_up_%02d", j), view)
	}

	// Ensure views differ across navigation
	b0, err := os.ReadFile(filepath.Join(tmp, "00_help_initial.txt"))
	req.NoError(err)
	bLast, err := os.ReadFile(filepath.Join(tmp, "09_help_up_03.txt"))
	req.NoError(err)
	req.NotEqual(string(b0), string(bLast))
}
