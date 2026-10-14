package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/xaxa/svyaz-cli/internal/llm"
)

type CoderAgent struct {
	client *llm.Client
}

func NewCoderAgent(client *llm.Client) *CoderAgent {
	return &CoderAgent{client: client}
}

const coderSystemPrompt = "You are the Coder Agent of svyaz-cli (Связь).\n" +
	"Your mission is to implement the required code changes strictly following the 'ponytail' directive:\n" +
	"1. YAGNI & EXTREME CONCISION: No overengineering, no unnecessary abstractions or redundant packages.\n" +
	"2. UNIFIED FORMAT: Output ONLY a valid git unified diff block (```diff ... ```).\n" +
	"3. DO NOT transcribe entire files. Generate only the delta/diff with sufficient context for 'git apply'.\n" +
	"4. FORMAT RULES FOR GIT APPLY:\n" +
	"   - The patch must begin with '--- a/path/to/file' and '+++ b/path/to/file'.\n" +
	"   - Every unchanged context line MUST start with a single space ' '.\n" +
	"   - Added lines start with '+', deleted lines start with '-'.\n" +
	"5. Expected format:\n" +
	"```diff\n" +
	"--- a/path/to/file.go\n" +
	"+++ b/path/to/file.go\n" +
	"@@ -10,4 +10,6 @@\n" +
	" existing context line\n" +
	"-removed line\n" +
	"+added line\n" +
	" existing context line\n" +
	"```\n"

// GeneratePatch receives the task and file context, producing a clean unified diff Patch
func (c *CoderAgent) GeneratePatch(
	ctx context.Context,
	task string,
	plan *Plan,
	fileContexts map[string]string,
	correctionFeedback string,
	onStream func(token string),
) (*Patch, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== APPROVED PLAN ===\n%s\n\n", plan.Summary))

	if len(fileContexts) > 0 {
		sb.WriteString("=== AFFECTED FILE CONTENTS ===\n")
		for path, content := range fileContexts {
			sb.WriteString(fmt.Sprintf("--- File: %s ---\n%s\n\n", path, content))
		}
	}

	if correctionFeedback != "" {
		sb.WriteString(fmt.Sprintf("=== PREVIOUS ERROR FEEDBACK (Dry-Run failed) ===\n%s\nFix the diff ensuring context lines match the exact file content.\n\n", correctionFeedback))
	}

	sb.WriteString(fmt.Sprintf("=== SPECIFIC TASK ===\n%s\nGenerate the unified diff now:\n", task))

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: llm.RoleSystem, Content: coderSystemPrompt},
			{Role: llm.RoleUser, Content: sb.String()},
		},
	}

	resp, err := c.client.ExecuteWithFallback(ctx, req, onStream)
	if err != nil {
		return nil, fmt.Errorf("coder execution failed: %w", err)
	}

	content := resp.Choices[0].Message.Content
	tokens := resp.Usage.TotalTokens
	if tokens == 0 {
		tokens = (len(sb.String()) + len(content)) / 4
		if tokens < 10 {
			tokens = 10
		}
	}

	filePath, diffText, addCount, delCount := ExtractDiff(content)

	if diffText == "" {
		return nil, fmt.Errorf("model did not produce a valid unified diff block")
	}

	return &Patch{
		FilePath:   filePath,
		Diff:       diffText,
		AddCount:   addCount,
		DelCount:   delCount,
		TokensUsed: tokens,
	}, nil
}

// ExtractDiff extracts unified diff block from markdown or raw response
func ExtractDiff(raw string) (filePath, diffText string, addCount, delCount int) {
	// Look inside ```diff or ```patch block first
	lines := strings.Split(raw, "\n")
	var diffLines []string
	inDiffBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```diff") || strings.HasPrefix(trimmed, "```patch") {
			inDiffBlock = true
			continue
		}
		if inDiffBlock {
			if strings.HasPrefix(trimmed, "```") {
				break
			}
			diffLines = append(diffLines, line)
		}
	}

	if len(diffLines) == 0 {
		// Fallback: look for lines starting with '--- a/' or 'diff --git'
		recording := false
		for _, line := range lines {
			if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "--- a/") || strings.HasPrefix(line, "--- /dev/null") {
				recording = true
			}
			if recording {
				if strings.HasPrefix(line, "```") {
					break
				}
				diffLines = append(diffLines, line)
			}
		}
	}

	if len(diffLines) == 0 {
		return "", "", 0, 0
	}

	// Normalize headers and ensure context lines start with space
	var normalized []string
	inHunk := false
	for _, l := range diffLines {
		if strings.HasPrefix(l, "--- ") && !strings.HasPrefix(l, "--- a/") && !strings.HasPrefix(l, "--- /dev/null") {
			target := strings.TrimSpace(strings.TrimPrefix(l, "--- "))
			l = "--- a/" + target
		} else if strings.HasPrefix(l, "+++ ") && !strings.HasPrefix(l, "+++ b/") && !strings.HasPrefix(l, "+++ /dev/null") {
			target := strings.TrimSpace(strings.TrimPrefix(l, "+++ "))
			l = "+++ b/" + target
		}

		if strings.HasPrefix(l, "@@") {
			inHunk = true
			normalized = append(normalized, l)
			continue
		}

		if inHunk {
			if strings.HasPrefix(l, "+") {
				addCount++
				normalized = append(normalized, l)
			} else if strings.HasPrefix(l, "-") {
				delCount++
				normalized = append(normalized, l)
			} else if strings.HasPrefix(l, " ") || strings.HasPrefix(l, `\`) {
				normalized = append(normalized, l)
			} else if l != "" {
				// Context line missing leading space
				normalized = append(normalized, " "+l)
			} else {
				// Empty line in hunk context
				normalized = append(normalized, " ")
			}
			continue
		}

		if strings.HasPrefix(l, "+++ b/") {
			filePath = strings.TrimPrefix(l, "+++ b/")
		} else if strings.HasPrefix(l, "--- a/") && filePath == "" {
			filePath = strings.TrimPrefix(l, "--- a/")
		}
		normalized = append(normalized, l)
	}

	diffText = strings.Join(normalized, "\n") + "\n"
	return filePath, diffText, addCount, delCount
}
