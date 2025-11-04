package integration

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/andreisuslov/cloud-sync/internal/ui"
)

// frameWriter captures terminal output and allows splitting into frames.
type frameWriter struct {
	mu   sync.Mutex
	buf  bytes.Buffer
}

func (w *frameWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	n, err := w.buf.Write(p)
	w.mu.Unlock()
	return n, err
}

func (w *frameWriter) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]byte(nil), w.buf.Bytes()...)
}

var (
	// Sequence used by bubbletea full-frame renderer to clear + home
	clearHome = []byte("\x1b[2J\x1b[H")
	ansiRe    = regexp.MustCompile("\\x1b\\[[0-9;]*[a-zA-Z]")
)

// splitFrames splits the captured output stream into full-frame snapshots.
func splitFrames(b []byte) [][]string {
	parts := bytes.Split(b, clearHome)
	frames := make([][]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		clean := ansiRe.ReplaceAll(p, nil)
		lines := strings.Split(string(clean), "\n")
		frames = append(frames, lines)
	}
	return frames
}

// countChangedLines compares two frames and returns the number of different lines.
func countChangedLines(a, b []string) int {
	max := len(a)
	if len(b) > max {
		max = len(b)
	}
	diff := 0
	for i := 0; i < max; i++ {
		var la, lb string
		if i < len(a) {
			la = a[i]
		}
		if i < len(b) {
			lb = b[i]
		}
		if la != lb {
			diff++
		}
	}
	return diff
}

func Test_MainMenu_RepaintsMultipleLinesOnUp(t *testing.T) {
	req := require.New(t)

	fw := &frameWriter{}
	p := tea.NewProgram(
		ui.NewModel(),
		tea.WithAltScreen(),
		tea.WithOutput(io.Writer(fw)),
	)

	// Run the program in background
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = p.Run()
	}()

	// Give it a moment to bootstrap
	time.Sleep(50 * time.Millisecond)

	// Set a small window to force scrolling
	p.Send(tea.WindowSizeMsg{Width: 60, Height: 12})
	time.Sleep(30 * time.Millisecond)

	// Scroll down many times to around 50%
	for i := 0; i < 12; i++ {
		p.Send(tea.KeyMsg{Type: tea.KeyDown})
		time.Sleep(20 * time.Millisecond)
	}

	// Capture frame before Up
	before := splitFrames(fw.Bytes())
	req.Greater(len(before), 0)
	prev := before[len(before)-1]

	// Press Up once
	p.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(40 * time.Millisecond)

	// Capture frame after Up
	after := splitFrames(fw.Bytes())
	req.Greater(len(after), 0)
	curr := after[len(after)-1]

	changed := countChangedLines(prev, curr)
	// Expect more than 1 line to change; top-row-only repaint would be 1 diff
	req.Greaterf(changed, 1, "expected multiple lines to change on Up, got %d", changed)

	// Quit program
	p.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	<-done
}

func Test_HelpViewport_RepaintsMultipleLinesOnUp(t *testing.T) {
	req := require.New(t)

	fw := &frameWriter{}
	p := tea.NewProgram(
		ui.NewModel(),
		tea.WithAltScreen(),
		tea.WithOutput(io.Writer(fw)),
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = p.Run()
	}()
	time.Sleep(50 * time.Millisecond)

	p.Send(tea.WindowSizeMsg{Width: 60, Height: 14})
	time.Sleep(30 * time.Millisecond)

	// Move to Help (menu item 6): five downs then enter
	for i := 0; i < 5; i++ {
		p.Send(tea.KeyMsg{Type: tea.KeyDown})
		time.Sleep(10 * time.Millisecond)
	}
	p.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(30 * time.Millisecond)

	// Scroll down a bunch
	for i := 0; i < 10; i++ {
		p.Send(tea.KeyMsg{Type: tea.KeyDown})
		time.Sleep(15 * time.Millisecond)
	}

	before := splitFrames(fw.Bytes())
	req.Greater(len(before), 0)
	prev := before[len(before)-1]

	// One Up
	p.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(40 * time.Millisecond)

	after := splitFrames(fw.Bytes())
	req.Greater(len(after), 0)
	curr := after[len(after)-1]

	changed := countChangedLines(prev, curr)
	req.Greaterf(changed, 1, "expected multiple lines to change on Up in help, got %d", changed)

	p.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	<-done
}
