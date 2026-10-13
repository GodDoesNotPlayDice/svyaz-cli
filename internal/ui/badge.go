package ui

import (
	"fmt"
)

type ActionStatus int

const (
	ActionPending ActionStatus = iota
	ActionSuccess
	ActionError
	ActionWarning
)

type ToolBadge struct {
	ToolName string
	Action   string
	Details  string
	Status   ActionStatus
}

// RenderToolBadge renders a discrete single-line tool execution indicator
func RenderToolBadge(badge ToolBadge) string {
	var dot string
	switch badge.Status {
	case ActionSuccess:
		dot = ActionSuccessDotStyle.Render("✓")
	case ActionWarning:
		dot = ActionWarningDotStyle.Render("▲")
	case ActionError:
		dot = ActionErrorDotStyle.Render("✖")
	default:
		dot = ActionDotStyle.Render("●")
	}

	badgeText := ActionBadgeStyle.Render(fmt.Sprintf("[%s:%s]", badge.ToolName, badge.Action))
	details := ActionTextStyle.Render(badge.Details)

	return fmt.Sprintf("  %s %s %s", dot, badgeText, details)
}
