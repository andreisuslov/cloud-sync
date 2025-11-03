package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andreisuslov/cloud-sync/internal/ui/styles"
)

// MainMenuView renders the main menu view
func MainMenuView(width, height int) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.RenderTitle("Cloud Sync - Backup Management"))
	b.WriteString("\n\n")
	b.WriteString(styles.RenderInfo("Select an option to continue:"))
	b.WriteString("\n\n")

	return b.String()
}

// MessageCmd returns a command that displays a message
func MessageCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// ViewHelper provides common view rendering utilities
type ViewHelper struct {
	Width  int
	Height int
}

// NewViewHelper creates a new view helper
func NewViewHelper(width, height int) *ViewHelper {
	return &ViewHelper{
		Width:  width,
		Height: height,
	}
}

// RenderHeader renders a view header
func (v *ViewHelper) RenderHeader(title, subtitle string) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.RenderTitle(title))
	b.WriteString("\n")
	if subtitle != "" {
		b.WriteString(styles.RenderSubtitle(subtitle))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

// RenderFooter renders a view footer with help text
func (v *ViewHelper) RenderFooter(helpText string) string {
	return fmt.Sprintf("\n\n%s", styles.RenderHelp(helpText))
}
