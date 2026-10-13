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
