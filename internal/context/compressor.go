package context

import (
	"fmt"
	"strings"
)

// CompressOutput truncates excessive terminal or compiler logs, keeping relevant head and tail lines
func CompressOutput(raw string, maxLines int) string {
	if maxLines <= 0 {
		maxLines = 60
	}

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= maxLines {
		return strings.TrimSpace(raw)
	}

	half := maxLines / 2
	head := lines[:half]
	tail := lines[len(lines)-half:]

	var sb strings.Builder
	sb.WriteString(strings.Join(head, "\n"))
	sb.WriteString(fmt.Sprintf("\n\n... [%d lines truncated by svyaz context compressor] ...\n\n", len(lines)-maxLines))
	sb.WriteString(strings.Join(tail, "\n"))
	return sb.String()
}
