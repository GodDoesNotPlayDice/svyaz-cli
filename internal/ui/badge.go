package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ActionStatus int

const (
	ActionPending ActionStatus = iota
	ActionSuccess
	ActionError
	ActionWarning
)

type ToolBadge struct {
	ToolName   string
	Action     string
	Details    string
	Status     ActionStatus
	Expandable bool
}

type ThoughtInfo struct {
	Duration string // e.g. "9s", "6s", "2s"
	Tokens   int    // e.g. 195, 188
	Summary  string // e.g. "Examining the terminal output. The '/h_' command suggests..."
}

// RenderToolBadge renders an Antigravity-style tool execution indicator:
// ● ToolName(Details) or ● ToolName(Details) (ctrl+o to expand)
func RenderToolBadge(badge ToolBadge) string {
	var dot string
	switch badge.Status {
	case ActionError:
		dot = ActionErrorDotStyle.Render("●")
	case ActionWarning:
		dot = ActionWarningDotStyle.Render("●")
	default:
		dot = ActionDotStyle.Render("●")
	}

	toolName := badge.ToolName
	if badge.Action != "" && !strings.Contains(toolName, "/") && toolName != "Bash" && toolName != "Edit" && toolName != "Git" {
		if toolName == "engram" {
			toolName = "engram/" + badge.Action
		}
	}

	nameStyled := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorTextPrimary)).
		Render(toolName)

	parenOpen := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Render("(")

	parenClose := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Render(")")

	argsText := badge.Details
	if argsText == "" {
		argsText = badge.Action
	}

	argsStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Render(argsText)

	expandStr := ""
	if badge.Expandable {
		expandStr = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Render(" (ctrl+o to expand)")
	}

	return fmt.Sprintf("%s %s%s%s%s%s", dot, nameStyled, parenOpen, argsStyled, parenClose, expandStr)
}

// RenderThought renders an Antigravity-style thought block:
// ▸ Thought for 9s, 195 tokens
//   Examining the terminal output. The `/h_` command suggests a help function or listin...
func RenderThought(thought ThoughtInfo, wrapWidth int) string {
	chevron := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccentBlue)).
		Render("▸")

	thoughtWord := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorTextPrimary)).
		Render("Thought")

	dur := thought.Duration
	if dur == "" {
		dur = "2s"
	}
	tok := thought.Tokens
	if tok <= 0 {
		tok = 120
	}

	meta := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Render(fmt.Sprintf("for %s, %d tokens", dur, tok))

	header := fmt.Sprintf("%s %s %s", chevron, thoughtWord, meta)

	if strings.TrimSpace(thought.Summary) == "" {
		return header
	}

	if wrapWidth < 20 {
		wrapWidth = 76
	}

	summaryStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color(ColorTextMuted)).
		Width(wrapWidth)

	return header + "\n" + summaryStyle.Render(thought.Summary)
}
