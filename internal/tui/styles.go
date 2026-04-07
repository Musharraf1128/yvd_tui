package tui

import "github.com/charmbracelet/lipgloss"

// Color palette (Retro Amber / Orange theme)
var (
	colorPrimary   = lipgloss.Color("#D97757") // Claude Code orange
	colorSecondary = lipgloss.Color("#E88158") // Slightly lighter orange
	colorAccent    = lipgloss.Color("#F59E0B") // Amber highlight
	colorSuccess   = lipgloss.Color("#10B981") // Emerald check
	colorError     = lipgloss.Color("#EF4444") // Red error
	colorDim       = lipgloss.Color("#6B7280") // Gray
	colorText      = lipgloss.Color("#F9FAFB") // Near-white
	colorSubtext   = lipgloss.Color("#9CA3AF") // Light gray
)

// Layout styles
var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			Background(colorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Italic(true)

	// Info display
	infoLabelStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Width(12)

	infoValueStyle = lipgloss.NewStyle().
			Foreground(colorText)

	// List items
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(colorSubtext)

	// Sections
	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Underline(true).
				MarginTop(1).
				MarginBottom(1)

	// Status
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Box for video info (no border anymore, just left padding)
	infoBoxStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorDim).
			MarginBottom(1)

	// Help
	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)

	// Cursor/pointer
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)
)
