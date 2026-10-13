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

	// Dispatch taskResultMsg with steps and tokens in Antigravity style
	res, _ = m.Update(taskResultMsg{
		Result: &agent.ExecutionResult{
			Intent:      agent.IntentChat,
			Response:    "Todo funcionando perfectamente.",
			TokensUsed:  250,
			DurationSec: 3,
			Thought:     "Clasificando intención de consulta...",
			Steps: []agent.ExecutionStep{
				{ToolName: "Planner", Action: "repo-map", Details: "AST y dependencias analizadas", Status: "success"},
				{ToolName: "Bash", Action: "test", Details: "go test -v ./...", Status: "success"},
			},
		},
	})
	m = res.(Model)

	view := m.View()

	// 1. Banner is at the top
	if !strings.Contains(view, "S V Y A Z") {
		t.Errorf("expected header banner at top")
	}

	// 2. Verbosity steps remain in Antigravity ● Tool(Details) format
	if !strings.Contains(view, "●") || !strings.Contains(view, "Planner(AST y dependencias analizadas)") {
		t.Errorf("expected Planner step to persist in Antigravity format, got:\n%s", view)
	}
	if !strings.Contains(view, "Bash(go test -v ./...)") {
		t.Errorf("expected Bash step to persist in Antigravity format, got:\n%s", view)
	}

	// 3. Tokens consumed and duration are visible in Antigravity ▸ Thought header
	if !strings.Contains(view, "Thought for 3s, 250 tokens") {
		t.Errorf("expected thought header with duration and tokens, got:\n%s", view)
	}

	// 4. Agent response is indented and has no svyaz: prefix
	if strings.Contains(view, "svyaz: ") {
		t.Errorf("expected agent response not to have svyaz: prefix")
	}
	if !strings.Contains(view, "Todo funcionando perfectamente.") {
		t.Errorf("expected agent response text to be present")
	}
}

func TestAntigravityVerbosityAndThoughts(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
	}

	m := NewModel(info)
	res, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = res.(Model)

	// Add thought followed by multiple compactly stacked tool steps
	m.addMessage(MessageItem{
		Role: "thought",
		Thought: &ThoughtInfo{
			Duration: "9s",
			Tokens:   195,
			Summary:  "Examining the terminal output. The `/h_` command suggests a help function...",
		},
	})
	m.addMessage(MessageItem{
		Role: "tool",
		Badge: &ToolBadge{
			ToolName: "Edit",
			Details:  "~/Documents/Github/svyaz-cli/internal/ui/header_test.go",
			Status:   ActionSuccess,
		},
	})
	m.addMessage(MessageItem{
		Role: "tool",
		Badge: &ToolBadge{
			ToolName:   "Bash",
			Details:    "go test -v -run TestSlashCommandAutocomplete ./internal/ui/...",
			Status:     ActionSuccess,
			Expandable: true,
		},
	})
	m.addMessage(MessageItem{
		Role: "tool",
		Badge: &ToolBadge{
			ToolName: "engram/mem_save",
			Details:  "Save slash command autocomplete decision",
			Status:   ActionSuccess,
		},
	})

	content := m.renderContent()

	// Check thought format
	if !strings.Contains(content, "▸") || !strings.Contains(content, "Thought") || !strings.Contains(content, "for 9s, 195 tokens") {
		t.Errorf("expected thought format ▸ Thought for 9s, 195 tokens, got:\n%s", content)
	}
	if !strings.Contains(content, "Examining the terminal output.") {
		t.Errorf("expected thought summary, got:\n%s", content)
	}

	// Check tool format
	if !strings.Contains(content, "●") || !strings.Contains(content, "Edit(~/Documents/Github/svyaz-cli/internal/ui/header_test.go)") {
		t.Errorf("expected Edit tool format, got:\n%s", content)
	}
	if !strings.Contains(content, "Bash(go test -v -run TestSlashCommandAutocomplete ./internal/ui/...)") || !strings.Contains(content, "(ctrl+o to expand)") {
		t.Errorf("expected Bash tool format with ctrl+o, got:\n%s", content)
	}
	if !strings.Contains(content, "engram/mem_save(Save slash command autocomplete decision)") {
		t.Errorf("expected engram tool format, got:\n%s", content)
	}

	// Check compact stacking: consecutive tool steps should have no blank lines separating them
	lines := strings.Split(content, "\n")
	foundEdit := -1
	foundBash := -1
	foundEngram := -1
	for idx, l := range lines {
		if strings.Contains(l, "Edit") && foundEdit == -1 {
			foundEdit = idx
		}
		if strings.Contains(l, "Bash") && foundBash == -1 {
			foundBash = idx
		}
		if strings.Contains(l, "engram/mem_save") && foundEngram == -1 {
			foundEngram = idx
		}
	}

	if foundEdit == -1 || foundBash == -1 || foundEngram == -1 {
		t.Fatalf("could not find all tool lines in:\n%s", content)
	}

	if !(foundEdit < foundBash && foundBash < foundEngram) {
		t.Fatalf("expected order edit < bash < engram, got edit=%d bash=%d engram=%d", foundEdit, foundBash, foundEngram)
	}

	for j := foundEdit; j <= foundEngram; j++ {
		if strings.TrimSpace(lines[j]) == "" {
			t.Errorf("expected no blank lines between consecutively stacked tools, but line %d was blank", j)
		}
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

func TestSlashCommandAutocomplete(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
	}

	m := NewModel(info)
	res, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = res.(Model)

	// 1. Test typing "/h" - suggests /help and shows navigation hints
	m.input.SetValue("/h")
	view := m.View()

	fmt.Println("--- AUTOCOMPLETE MENU FOR /h ---")
	fmt.Println(view)

	if !strings.Contains(view, "/help") {
		t.Errorf("expected autocomplete menu to contain /help, got:\n%s", view)
	}
	if !strings.Contains(view, "Navigate") || !strings.Contains(view, "enter Select") || !strings.Contains(view, "tab complete") {
		t.Errorf("expected autocomplete navigation hint, got:\n%s", view)
	}
	if !strings.Contains(view, "esc para salir de ese pequeño menú") {
		t.Errorf("expected esc exit hint, got:\n%s", view)
	}

	// 2. Test typing "/" - lists all commands and allows arrow navigation
	m.input.SetValue("/")
	m.autocompleteDismissed = false
	m.autocompleteIndex = 0

	fmt.Println("--- AUTOCOMPLETE MENU FOR / ---")
	fmt.Println(m.View())

	// Press KeyDown: should select second command (/clear)
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.autocompleteIndex != 1 {
		t.Errorf("expected autocompleteIndex to be 1 after KeyDown, got %d", m.autocompleteIndex)
	}

	// 3. Test Tab key completion
	// Set to "/h" and press Tab: should complete to "/help "
	m.input.SetValue("/h")
	m.autocompleteDismissed = false
	m.autocompleteIndex = 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(Model)

	if !strings.HasPrefix(m.input.Value(), "/help") {
		t.Errorf("expected input to autocomplete to /help, got %q", m.input.Value())
	}

	// 4. Test Esc key dismisses menu while keeping text
	m.input.SetValue("/h")
	m.autocompleteDismissed = false
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)

	if !m.autocompleteDismissed {
		t.Errorf("expected autocomplete to be dismissed on Esc")
	}
	viewAfterEsc := m.View()
	if strings.Contains(viewAfterEsc, "Navigate · enter Select") {
		t.Errorf("expected menu to be hidden after Esc, but was present")
	}
	if m.input.Value() != "/h" {
		t.Errorf("expected input to retain '/h', got %q", m.input.Value())
	}

	// 5. Test Enter key selects and executes the command
	m.input.SetValue("/h")
	m.autocompleteDismissed = false
	m.autocompleteIndex = 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	// Since /help was selected, messages should contain help text
	viewAfterEnter := m.View()
	if !strings.Contains(viewAfterEnter, "Atajos y comandos disponibles") {
		t.Errorf("expected /help command to execute on Enter, got:\n%s", viewAfterEnter)
	}
}
