package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type KeyHint struct {
	Key  string
	Desc string
}

// RenderFooter renders the bottom bar with '? for shortcuts' on the left and the active model on the right
func RenderFooter(modelName string, width int) string {
	if modelName == "" {
		modelName = "deepseek-coder"
	}

	left := FooterHintStyle.Render("? for shortcuts")
	right := FooterStatusStyle.Render(modelName)

	paddingX := 1
	availableWidth := width - (paddingX * 2)

	gap := availableWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	content := left + strings.Repeat(" ", gap) + right
	return FooterContainerStyle.Width(width).Render(content)
}

// RenderStatusBar fallback for structured hints
func RenderStatusBar(hints []KeyHint, width int) string {
	return RenderFooter("deepseek-coder", width)
}
