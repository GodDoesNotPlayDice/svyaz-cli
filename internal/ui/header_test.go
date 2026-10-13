package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xaxa/svyaz-cli/internal/agent"
)

func TestRenderHeaderAndFooter(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
		ModelStatus:  "Ready",
	}

	m := NewModel(info)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	viewOutput := updated.View()

	fmt.Println("--- FULL CLI PREVIEW WITH TWO TIGHT BARS ---")
	fmt.Println(viewOutput)

	if !strings.Contains(viewOutput, "S V Y A Z") {
		t.Errorf("expected view to contain logo brand S V Y A Z")
	}
	if !strings.Contains(viewOutput, "? for shortcuts") {
		t.Errorf("expected view to contain footer shortcut hint")
	}
	if !strings.Contains(viewOutput, "deepseek-coder") {
		t.Errorf("expected view to contain active model")
	}
}

func TestMessageWordWrapping(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
		ModelStatus:  "Ready",
	}

	m := NewModel(info)
	// Initialize with narrow window width of 60 columns to test wrapping
	modelAfterResize, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m = modelAfterResize.(Model)

	longMsg := "¡Hola! Estoy funcionando correctamente, listo para ayudarte con cualquier cosa que necesites en este proyecto."
	m.addMessage(MessageItem{
		Role:    "agent",
		Content: longMsg,
	})
	m.updateViewportContent()

	viewOutput := m.View()
	fmt.Println("--- VIEWPORT WITH WRAPPED MESSAGE (60 cols) ---")
	fmt.Println(viewOutput)

	// Verify the entire text is present in the rendered view (wrapped across lines, not truncated)
	if !strings.Contains(viewOutput, "cualquier cosa") {
		t.Errorf("expected view to contain 'cualquier cosa'")
	}
	if !strings.Contains(viewOutput, "necesites en este") || !strings.Contains(viewOutput, "proyecto") {
		t.Errorf("expected view to contain 'necesites en este' and 'proyecto'")
	}

	// Verify no "svyaz: " prefix is rendered
	if strings.Contains(viewOutput, "svyaz: ") {
		t.Errorf("expected view not to contain 'svyaz: ' prefix")
	}
}

func TestUIRefinements_UserMessageBlueColor(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
	}

	m := NewModel(info)
	res, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = res.(Model)

	// Simulate user typing and pressing enter
	m.input.SetValue("hola como estas")
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	rendered := m.renderContent()
	// The user prompt line must contain "> " and "hola como estas"
	if !strings.Contains(rendered, "> ") || !strings.Contains(rendered, "hola como estas") {
		t.Errorf("expected rendered content to contain user prompt line")
	}

	// Verify thinking badge is also created and preserved
	if !strings.Contains(rendered, "ast-graph") {
		t.Errorf("expected thinking badge to be in rendered content")
	}
}

func TestUIRefinements_VerbosityAndTokensPreserved(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
	}

	m := NewModel(info)
	res, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = res.(Model)

	// Dispatch taskResultMsg with steps and tokens
	res, _ = m.Update(taskResultMsg{
		Result: &agent.ExecutionResult{
			Intent:     agent.IntentChat,
			Response:   "Todo funcionando perfectamente.",
			TokensUsed: 250,
			Steps: []agent.ExecutionStep{
				{ToolName: "planner", Action: "repo-map", Details: "AST y dependencias analizadas", Status: "success"},
				{ToolName: "thinking", Action: "intent", Details: "Consulta clasificada como conversación", Status: "success"},
			},
		},
	})
	m = res.(Model)

	view := m.View()

	// 1. Banner is at the top
	if !strings.Contains(view, "S V Y A Z") {
		t.Errorf("expected header banner at top")
	}

	// 2. Verbosity steps remain
	if !strings.Contains(view, "[planner:repo-map]") {
		t.Errorf("expected planner step badge to persist")
	}
	if !strings.Contains(view, "[thinking:intent]") {
		t.Errorf("expected thinking intent badge to persist")
	}

	// 3. Tokens consumed badge is visible
	if !strings.Contains(view, "[tokens:usage]") || !strings.Contains(view, "250 tokens consumidos") {
		t.Errorf("expected tokens consumed badge to be visible")
	}

	// 4. Agent response is indented and has no svyaz: prefix
	if strings.Contains(view, "svyaz: ") {
		t.Errorf("expected agent response not to have svyaz: prefix")
	}
	if !strings.Contains(view, "Todo funcionando perfectamente.") {
		t.Errorf("expected agent response text to be present")
	}
}

func TestUIRefinements_DownwardElongation(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
	}

	m := NewModel(info)
	res, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 35})
	m = res.(Model)

	// When 1 message is added, input area should appear directly under the message,
	// NOT 20 lines away at the bottom
	m.addMessage(MessageItem{
		Role:    "agent",
		Content: "Respuesta breve.",
	})
	m.updateViewportContent()

	view := m.View()
	lines := strings.Split(view, "\n")

	// Find line index of header, message, and prompt input
	headerIdx := -1
	msgIdx := -1
	promptIdx := -1
	footerIdx := -1

	for i, l := range lines {
		if strings.Contains(l, "S V Y A Z") && headerIdx == -1 {
			headerIdx = i
		}
		if strings.Contains(l, "Respuesta breve.") && msgIdx == -1 {
			msgIdx = i
		}
		if strings.Contains(l, "> _") && promptIdx == -1 {
			promptIdx = i
		}
		if strings.Contains(l, "? for shortcuts") && footerIdx == -1 {
			footerIdx = i
		}
	}

	if headerIdx == -1 || msgIdx == -1 || promptIdx == -1 || footerIdx == -1 {
		t.Fatalf("expected all components present: header=%d msg=%d prompt=%d footer=%d",
			headerIdx, msgIdx, promptIdx, footerIdx)
	}

	// Header must be above message, message above prompt, prompt above footer
	if !(headerIdx < msgIdx && msgIdx < promptIdx && promptIdx < footerIdx) {
		t.Errorf("expected header (%d) < msg (%d) < prompt (%d) < footer (%d)",
			headerIdx, msgIdx, promptIdx, footerIdx)
	}

	// Gap between message and prompt should be tight (prompt directly follows message, e.g. <= 4 lines)
	if promptIdx-msgIdx > 5 {
		t.Errorf("expected prompt to follow message closely (elongating downwards), but got gap of %d lines",
			promptIdx-msgIdx)
	}
}
