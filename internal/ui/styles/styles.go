package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#00ADD8") // Go blue
	SecondaryColor = lipgloss.Color("#5DC9E2")
	SuccessColor   = lipgloss.Color("#00D787")
	WarningColor   = lipgloss.Color("#FFA500")
	ErrorColor     = lipgloss.Color("#FF5555")
	MutedColor     = lipgloss.Color("#888888")
	BorderColor    = lipgloss.Color("#3C3C3C")

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			PaddingLeft(2).
			PaddingBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			PaddingLeft(2)

	// Menu item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(PrimaryColor).
				PaddingLeft(2).
				PaddingRight(2).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC")).
				PaddingLeft(2).
				PaddingRight(2)

	DimmedItemStyle = lipgloss.NewStyle().
				Foreground(MutedColor).
				PaddingLeft(2).
				PaddingRight(2)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			PaddingTop(1).
			PaddingLeft(2)

	// Progress bar style
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor)

	// Log viewer styles
	LogLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#DDDDDD"))

	LogErrorStyle = lipgloss.NewStyle().
				Foreground(ErrorColor).
				Bold(true)

	LogWarningStyle = lipgloss.NewStyle().
				Foreground(WarningColor)

	LogInfoStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor)

	// Viewport style
	ViewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	// Highlight style
	HighlightStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// Muted style
	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)
	
	// Input field styles
	FocusedStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor)
	
	NoStyle = lipgloss.NewStyle()
)

// RenderTitle renders a styled title
func RenderTitle(title string) string {
	return TitleStyle.Render(title)
}

// RenderSubtitle renders a styled subtitle
func RenderSubtitle(subtitle string) string {
	return SubtitleStyle.Render(subtitle)
}

// RenderSuccess renders success text
func RenderSuccess(text string) string {
	return SuccessStyle.Render(text)
}

// RenderError renders error text
func RenderError(text string) string {
	return ErrorStyle.Render(text)
}

// RenderWarning renders warning text
func RenderWarning(text string) string {
	return WarningStyle.Render(text)
}

// RenderInfo renders info text
func RenderInfo(text string) string {
	return InfoStyle.Render(text)
}

// RenderBox renders content in a styled box
func RenderBox(content string) string {
	return BoxStyle.Render(content)
}

// RenderHelp renders help text
func RenderHelp(text string) string {
	return HelpStyle.Render(text)
}

// RenderHighlight renders highlighted text
func RenderHighlight(text string) string {
	return HighlightStyle.Render(text)
}

// RenderMuted renders muted text
func RenderMuted(text string) string {
	return MutedStyle.Render(text)
}