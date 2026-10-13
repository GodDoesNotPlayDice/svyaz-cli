package ui

import "github.com/charmbracelet/lipgloss"

// Color Palette (Antigravity-inspired sober minimalism)
const (
	ColorBgBase       = "#18181B" // Zinc 900
	ColorTextPrimary  = "#E4E4E7" // Zinc 200
	ColorTextMuted    = "#71717A" // Zinc 500
	ColorTextDim      = "#52525B" // Zinc 600
	ColorBorder       = "#27272A" // Zinc 800
	ColorBorderActive = "#3F3F46" // Zinc 700

	ColorAccentBlue  = "#38BDF8" // Sky 400 (Sober brand accent)
	ColorAccentCyan  = "#06B6D4" // Cyan 500
	ColorSuccess     = "#22C55E" // Emerald 500
	ColorWarning     = "#EAB308" // Amber 500
	ColorDanger      = "#EF4444" // Red 500
	ColorDiffAddBg   = "#052e16" // Dark green background for diffs
	ColorDiffDelBg   = "#450a0a" // Dark red background for diffs
)

// UI Styles
var (
	// Brand and Title
	BrandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorAccentBlue))

	VersionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim))

	// Header Styles
	HeaderContainerStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color(ColorBorder))

	HeaderItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextMuted))

	HeaderItemValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextPrimary))

	HeaderAccentValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorAccentBlue))

	HeaderSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorBorderActive)).
				SetString(" │ ")

	// Prompt Box Styles
	PromptContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorderActive)).
				Padding(0, 1)

	PromptPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorAccentBlue))

	PromptPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim))

	// Status Bar Styles
	StatusBarContainerStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.NormalBorder(), true, false, false, false).
				BorderForeground(lipgloss.Color(ColorBorder))

	StatusBarKeyStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorTextPrimary))

	StatusBarDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim))

	StatusBarDotStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorBorderActive)).
				SetString("  •  ")

	// Message / Chat History Styles
	UserMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextPrimary)).
				Bold(true)

	UserPrefixStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccentBlue)).
			Bold(true)

	AgentMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextPrimary))

	AgentPrefixStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorAccentCyan)).
				Bold(true)

	// Discrete Tool Action Badges
	ActionDotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccentBlue))

	ActionSuccessDotStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSuccess))

	ActionErrorDotStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDanger))

	ActionBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorAccentBlue)).
				Bold(true)

	ActionTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextMuted))

	// Diff View Styles
	DiffBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorderActive)).
			Padding(0, 1).
			Margin(1, 0)

	DiffHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextMuted)).
			Bold(true)

	DiffHunkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccentCyan))

	DiffAddStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess))

	DiffDelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDanger))

	DiffPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextPrimary)).
			Bold(true)

	DiffChoiceSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorAccentBlue)).
				Bold(true)

	DiffChoiceDimStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim))
)
