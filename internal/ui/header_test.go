package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
}
