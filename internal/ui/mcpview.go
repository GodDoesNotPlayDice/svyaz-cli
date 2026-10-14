package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xaxa/svyaz-cli/internal/mcp"
)

// RenderMCPView renders the MCP Servers list matching Antigravity's interface
func RenderMCPView(servers []mcp.MCPServerInfo, location string, selectedIndex int, expandedServer string, width int) string {
	if width < 40 {
		width = 80
	}

	var sb strings.Builder

	// Header: MCP Servers
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorTextPrimary))
	sb.WriteString(titleStyle.Render("MCP Servers") + "\n\n")

	// Plugins location label: Plugins (~/.gemini/config/plugins)
	pluginsLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextMuted))
	sb.WriteString(pluginsLabelStyle.Render(fmt.Sprintf("Plugins (%s)", location)) + "\n")

	if len(servers) == 0 {
		sb.WriteString("   " + lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim)).Render("No MCP servers found.") + "\n\n")
	} else {
		for i, srv := range servers {
			isSelected := i == selectedIndex
			renderServerItem(&sb, srv, isSelected, expandedServer == srv.Name, width)
		}
	}

	sb.WriteString("\n")

	// Footer: Keyboard: ↑/↓ Navigate  enter Actions  esc Exit
	kbLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim)).Render("Keyboard: ")
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBlue)).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim))

	footer := kbLabel +
		keyStyle.Render("↑/↓") + descStyle.Render(" Navigate  ") +
		keyStyle.Render("enter") + descStyle.Render(" Actions  ") +
		keyStyle.Render("esc") + descStyle.Render(" Exit")

	sb.WriteString(footer)
	return sb.String()
}

func renderServerItem(sb *strings.Builder, srv mcp.MCPServerInfo, isSelected bool, isExpanded bool, termWidth int) {
	// Cursor prefix: ">  " or "   "
	cursor := "   "
	if isSelected {
		cursor = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBlue)).Bold(true).Render(">  ")
	}

	// Status checkmark: "✓ " in green
	checkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess))
	if srv.Status == "✗" {
		checkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDanger))
	}
	checkMark := checkStyle.Render("✓ ")

	// Server Name
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorTextPrimary))
	if isSelected {
		nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorAccentBlue))
	}
	nameStr := nameStyle.Render(srv.Name)

	// Auth Needed badge if applicable
	authStr := ""
	if srv.AuthNeeded {
		authStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Bold(true)
		authStr = "  " + authStyle.Render("[Auth Needed]")
	}

	// Tools label
	toolsLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim)).Render("  Tools: ")

	// Calculate prefix visible width for wrapping indentation
	// prefix: "   ✓ ServerName  [Auth Needed]  Tools: "
	prefixPlain := fmt.Sprintf("   ✓ %s", srv.Name)
	if srv.AuthNeeded {
		prefixPlain += "  [Auth Needed]"
	}
	prefixPlain += "  Tools: "
	indentLen := len(prefixPlain)

	// If expanded, show all tools or detailed card
	if isExpanded && len(srv.Tools) > 0 {
		sb.WriteString(fmt.Sprintf("%s%s%s%s%s\n", cursor, checkMark, nameStr, authStr, toolsLabel))
		// Render expanded tools block
		expandIndent := strings.Repeat(" ", indentLen)
		toolsText := strings.Join(srv.Tools, ", ")
		wrapped := wrapText(toolsText, termWidth-indentLen-2)
		for _, line := range wrapped {
			sb.WriteString(expandIndent + lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextPrimary)).Render(line) + "\n")
		}
		return
	}

	// Format tools string: first up to 5 tools, then +N more
	toolsSummary := formatToolsSummary(srv.Tools)
	wrappedTools := wrapText(toolsSummary, termWidth-indentLen-2)

	if len(wrappedTools) == 0 {
		sb.WriteString(fmt.Sprintf("%s%s%s%s%s\n", cursor, checkMark, nameStr, authStr, toolsLabel))
		return
	}

	firstLine := wrappedTools[0]
	sb.WriteString(fmt.Sprintf("%s%s%s%s%s%s\n", cursor, checkMark, nameStr, authStr, toolsLabel, formatToolsStyling(firstLine)))

	indent := strings.Repeat(" ", indentLen)
	for i := 1; i < len(wrappedTools); i++ {
		sb.WriteString(indent + formatToolsStyling(wrappedTools[i]) + "\n")
	}
}

func formatToolsSummary(tools []string) string {
	if len(tools) == 0 {
		return "none"
	}

	maxDisplay := 5
	if len(tools) <= maxDisplay {
		return strings.Join(tools, ", ")
	}

	displayed := tools[:maxDisplay]
	remaining := len(tools) - maxDisplay
	return fmt.Sprintf("%s, +%d more", strings.Join(displayed, ", "), remaining)
}

func formatToolsStyling(line string) string {
	toolsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextMuted))
	moreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextDim)).Italic(true)

	if idx := strings.Index(line, "+"); idx != -1 {
		part1 := line[:idx]
		part2 := line[idx:]
		return toolsStyle.Render(part1) + moreStyle.Render(part2)
	}

	return toolsStyle.Render(line)
}

func wrapText(text string, maxWidth int) []string {
	if maxWidth < 20 {
		maxWidth = 70
	}

	words := strings.Split(text, " ")
	var lines []string
	var current strings.Builder

	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
		} else if current.Len()+1+len(word) <= maxWidth {
			current.WriteString(" " + word)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}
