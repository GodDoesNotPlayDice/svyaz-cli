package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type KeyHint struct {
	Key  string
	Desc string
}

// RenderFooter renders the bottom bar with '? for shortcuts' on the left and model status on the right
func RenderFooter(modelStatus string, width int) string {
	if modelStatus == "" {
		modelStatus = "Ready"
	}

	left := FooterHintStyle.Render("? for shortcuts")
	dot := FooterDotReadyStyle.Render("●")
	right := fmt.Sprintf("%s %s", dot, FooterStatusStyle.Render(modelStatus))

	paddingX := 2
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
	return RenderFooter("Ready", width)
}
