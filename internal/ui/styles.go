package ui

import "github.com/charmbracelet/lipgloss"

// Color Palette (Antigravity-inspired sober minimalism)
const (
	ColorBgBase       = "#18181B" // Zinc 900
	ColorTextPrimary  = "#E4E4E7" // Zinc 200
	ColorTextMuted    = "#71717A" // Zinc 500
	ColorTextDim      = "#52525B" // Zinc 600
	ColorBorder       = "#27272A" // Zinc 800
	ColorBorderLight  = "#46474A" // Gris tenue y fino exacto (#46474A)
	ColorBorderActive = "#3F3F46" // Zinc 700

	ColorAccentBlue = "#38BDF8" // Sky 400 (Sober brand accent)
	ColorAccentCyan = "#06B6D4" // Cyan 500
	ColorSuccess    = "#22C55E" // Emerald 500
	ColorWarning    = "#EAB308" // Amber 500
	ColorDanger     = "#EF4444" // Red 500
	ColorDiffAddBg  = "#052e16" // Dark green background for diffs
	ColorDiffDelBg  = "#450a0a" // Dark red background for diffs
)

// UI Styles
var (
	// Antigravity-style Logo & Metadata
	LogoChainStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccentBlue))

	LogoTextStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorAccentBlue))

	MetaTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorTextPrimary))

	MetaVersionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim))

	MetaLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextMuted))

	MetaModelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorAccentBlue))

	MetaPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim))

	SeparatorLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorBorderLight)).
				Bold(false)

	// Brand and Title
	BrandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorAccentBlue))

	VersionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim))

	// Prompt Styles
	PromptPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorAccentBlue))

	PromptInputTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextPrimary))

	PromptPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim))

	// Status Bar / Footer Styles
	FooterContainerStyle = lipgloss.NewStyle().
				Padding(0, 1)

	FooterHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim))

	FooterStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextMuted))

	FooterDotReadyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSuccess))

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
