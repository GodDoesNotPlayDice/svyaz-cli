package ui

import (
	"fmt"
	"strings"
)

type DiffPatch struct {
	FilePath string
	DiffText string
	AddCount int
	DelCount int
}

// RenderDiffView renders a sober unified diff preview with confirmation buttons
func RenderDiffView(patch DiffPatch, width int) string {
	header := DiffHeaderStyle.Render(fmt.Sprintf("Patch: %s (+%d -%d)", patch.FilePath, patch.AddCount, patch.DelCount))

	var lines []string
	lines = append(lines, header, "")

	rawLines := strings.Split(patch.DiffText, "\n")
	maxLines := 15
	if len(rawLines) > maxLines {
		rawLines = rawLines[:maxLines]
		rawLines = append(rawLines, "... (truncated)")
	}

	for _, line := range rawLines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			lines = append(lines, DiffAddStyle.Render(line))
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			lines = append(lines, DiffDelStyle.Render(line))
		} else if strings.HasPrefix(line, "@@") {
			lines = append(lines, DiffHunkStyle.Render(line))
		} else {
			lines = append(lines, ActionTextStyle.Render(line))
		}
	}

	// Decision prompt
	decisionPrompt := fmt.Sprintf(
		"\n%s  %s  %s",
		DiffChoiceSelectedStyle.Render("[y] Apply"),
		DiffChoiceDimStyle.Render("[e] Edit"),
		DiffChoiceDimStyle.Render("[d] Discard"),
	)
	lines = append(lines, decisionPrompt)

	body := strings.Join(lines, "\n")

	box := DiffBoxStyle
	if width > 4 {
		box = box.Width(width - 4)
	}
	return box.Render(body)
}
