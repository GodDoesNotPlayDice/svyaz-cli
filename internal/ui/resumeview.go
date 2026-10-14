package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xaxa/svyaz-cli/internal/session"
)

// RenderResumeView renders the saved sessions list for /resume
func RenderResumeView(sessions []session.Session, selectedIndex int, width int) string {
	if width < 40 {
		width = 80
	}

	var sb strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorTextPrimary))
	sb.WriteString(titleStyle.Render("Chat Sessions") + "\n\n")

	dir, _ := session.GetStorageDir()
	homePrefix := ""
	if strings.HasPrefix(dir, "/home/") {
		parts := strings.SplitN(dir, "/", 4)
		if len(parts) >= 4 {
			homePrefix = "~/" + parts[3]
		}
	}
	if homePrefix == "" {
		homePrefix = dir
	}

	subLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextMuted))
	sb.WriteString(subLabelStyle.Render(fmt.Sprintf("Saved in %s", homePrefix)) + "\n\n")

	if len(sessions) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim))
		sb.WriteString("   " + dimStyle.Render("No previous chat sessions found. Type a prompt or run /help to start.") + "\n\n")
	} else {
		for i, sess := range sessions {
			isSelected := i == selectedIndex

			cursor := "   "
			if isSelected {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBlue)).Bold(true).Render(">  ")
			}

			// Date formatted: 2006-01-02 15:04
			dateStr := sess.UpdatedAt.Format("2006-01-02 15:04")
			dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim))

			// Count of messages
			countStr := fmt.Sprintf("[%d msgs]", len(sess.Messages))
			countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentCyan))

			// Title
			title := sess.Title
			if title == "" {
				title = "Session " + sess.ID
			}
			titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextPrimary))
			if isSelected {
				titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBlue)).Bold(true)
			}

			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim))
			idStr := idStyle.Render(fmt.Sprintf("(%s)", sess.ID))

			sb.WriteString(fmt.Sprintf("%s%s  %s  %s  %s\n",
				cursor,
				dateStyle.Render(dateStr),
				countStyle.Render(countStr),
				titleStyle.Render(title),
				idStr,
			))
		}
	}

	sb.WriteString("\n")

	// Keyboard hint
	kbLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim)).Render("Keyboard: ")
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBlue)).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim))

	footer := kbLabel +
		keyStyle.Render("↑/↓") + descStyle.Render(" Navigate  ") +
		keyStyle.Render("enter") + descStyle.Render(" Resume  ") +
		keyStyle.Render("d") + descStyle.Render(" Delete  ") +
		keyStyle.Render("esc") + descStyle.Render(" Exit")

	sb.WriteString(footer)
	return sb.String()
}
