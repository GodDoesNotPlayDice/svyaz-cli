package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/xaxa/svyaz-cli/internal/llm"
)

type PlannerAgent struct {
	client *llm.Client
}

func NewPlannerAgent(client *llm.Client) *PlannerAgent {
	return &PlannerAgent{client: client}
}

const plannerSystemPrompt = `You are svyaz-cli (Связь), a high-performance, minimalist agentic coding assistant in Go.
Your first responsibility is to intelligently determine the user's INTENT:

1. [CHAT]:
- Applies when the user greets ("hello", "hi", "hola"), asks conceptual or theoretical questions, inquires about what the project does, requests explanations of code or commands, or has a conversation.
- Does NOT require modifying, creating, or deleting any file on disk.
- RULE: Begin your response EXACTLY with "[CHAT]" followed by your direct, natural, concise, and helpful answer.
- LANGUAGE RULE: Respond in the same language the user addressed you in (e.g. if the user speaks Spanish, answer in Spanish; if English, answer in English). NEVER invent fictional files (like hello.go) for a greeting or question.

2. [CODE_CHANGE]:
- Applies ONLY when the user explicitly asks to create a feature, write code, fix a bug, refactor, or edit real files in the repository.
- CONDUCT RULES (Layer 1 Grounding Harness):
  1. ZERO DIRECT CODE: Only plan, do not generate the full implementation here.
  2. REALITY GROUNDING: Strictly forbidden from inventing dependencies or libraries outside the Repo Map.
  3. MINIMALISM (YAGNI): Identify ONLY the exact files and symbols to edit.
- RULE: Begin your response EXACTLY with "[CODE_CHANGE]" followed by the structured plan:
[CODE_CHANGE]
### Plan Summary
[1-2 lines]

### Files and Symbols
- path/to/file.ext: [description]

### Execution Steps
1. [Step 1]
2. [Step 2]
`

// PlanTask analyzes user request against repo map and determines intent (chat vs code change plan)
func (p *PlannerAgent) PlanTask(
	ctx context.Context,
	task string,
	repoMap string,
	onStream func(token string),
) (*PlannerResult, error) {
	prompt := fmt.Sprintf("%s\n\n=== MENSAJE DEL USUARIO ===\n%s\n", repoMap, task)

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: llm.RoleSystem, Content: plannerSystemPrompt},
			{Role: llm.RoleUser, Content: prompt},
		},
	}

	resp, err := p.client.ExecuteWithFallback(ctx, req, onStream)
	if err != nil {
		return nil, fmt.Errorf("planner execution failed: %w", err)
	}

	content := resp.Choices[0].Message.Content
	tokens := resp.Usage.TotalTokens
	if tokens == 0 {
		tokens = (len(prompt) + len(content)) / 4
		if tokens < 10 {
			tokens = 10
		}
	}
	res := parsePlannerResult(content)
	res.TokensUsed = tokens
	return res, nil
}

func parsePlannerResult(content string) *PlannerResult {
	trimmed := strings.TrimSpace(content)

	if strings.HasPrefix(trimmed, "[CHAT]") {
		chatContent := strings.TrimSpace(strings.TrimPrefix(trimmed, "[CHAT]"))
		return &PlannerResult{
			Intent:  IntentChat,
			Content: chatContent,
			Plan:    nil,
		}
	}

	if strings.HasPrefix(trimmed, "[CODE_CHANGE]") {
		planContent := strings.TrimSpace(strings.TrimPrefix(trimmed, "[CODE_CHANGE]"))
		return &PlannerResult{
			Intent:  IntentCodeChange,
			Content: planContent,
			Plan:    parsePlanContent(planContent),
		}
	}

	// Fallback heuristic: check if content contains plan sections
	if strings.Contains(trimmed, "### Files") || strings.Contains(trimmed, "### Archivos") ||
		strings.Contains(trimmed, "### Steps") || strings.Contains(trimmed, "### Pasos") {
		return &PlannerResult{
			Intent:  IntentCodeChange,
			Content: trimmed,
			Plan:    parsePlanContent(trimmed),
		}
	}

	// Default to chat if no code changes planned
	return &PlannerResult{
		Intent:  IntentChat,
		Content: trimmed,
		Plan:    nil,
	}
}

func parsePlanContent(content string) *Plan {
	lines := strings.Split(content, "\n")
	plan := &Plan{
		Summary: "",
		Steps:   []PlanStep{},
		Files:   []string{},
	}

	var currentSection string
	var summaryLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "### plan summary") || strings.HasPrefix(lower, "### resumen") {
			currentSection = "summary"
			continue
		} else if strings.HasPrefix(lower, "### files") || strings.HasPrefix(lower, "### archivos") {
			currentSection = "files"
			continue
		} else if strings.HasPrefix(lower, "### execution steps") || strings.HasPrefix(lower, "### steps") || strings.HasPrefix(lower, "### pasos") {
			currentSection = "steps"
			continue
		}

		switch currentSection {
		case "summary":
			if trimmed != "" {
				summaryLines = append(summaryLines, trimmed)
			}
		case "files":
			if trimmed != "" && (strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*")) {
				item := strings.TrimLeft(trimmed, "-* ")
				item = strings.Trim(item, "`")
				parts := strings.SplitN(item, ":", 2)
				candidate := strings.TrimSpace(parts[0])
				candidate = strings.Trim(candidate, "`\"' ")
				// If there's a symbol like (func main), take the path
				fields := strings.Fields(candidate)
				if len(fields) > 0 {
					filePath := fields[0]
					if filePath != "" {
						plan.Files = append(plan.Files, filePath)
					}
				}
			}
		case "steps":
			if strings.HasPrefix(trimmed, "1.") || strings.HasPrefix(trimmed, "2.") ||
				strings.HasPrefix(trimmed, "3.") || strings.HasPrefix(trimmed, "4.") ||
				strings.HasPrefix(trimmed, "5.") || strings.HasPrefix(trimmed, "-") {
				desc := strings.TrimLeft(trimmed, "0123456789.- ")
				plan.Steps = append(plan.Steps, PlanStep{
					Description: desc,
					Files:       plan.Files,
				})
			}
		}
	}

	if len(summaryLines) > 0 {
		plan.Summary = strings.Join(summaryLines, " ")
	} else {
		plan.Summary = strings.TrimSpace(content)
	}

	return plan
}
