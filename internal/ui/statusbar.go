package ui

import (
	"github.com/charmbracelet/lipgloss"
)

type KeyHint struct {
	Key  string
	Desc string
}

// RenderStatusBar renders a minimalist status bar at the bottom
func RenderStatusBar(hints []KeyHint, width int) string {
	if len(hints) == 0 {
		hints = []KeyHint{
			{Key: "Enter", Desc: "enviar"},
			{Key: "Esc", Desc: "cancelar"},
			{Key: "Tab", Desc: "autocompletar"},
			{Key: "/help", Desc: "comandos"},
			{Key: "Ctrl+C", Desc: "salir"},
		}
	}

	var items []string
	dot := StatusBarDotStyle.String()

	for i, h := range hints {
		rendered := StatusBarKeyStyle.Render(h.Key) + " " + StatusBarDescStyle.Render(h.Desc)
		items = append(items, rendered)
		if i < len(hints)-1 {
			items = append(items, dot)
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Center, items...)
	if width > 0 {
		return StatusBarContainerStyle.Width(width).Render(content)
	}
	return StatusBarContainerStyle.Render(content)
}
