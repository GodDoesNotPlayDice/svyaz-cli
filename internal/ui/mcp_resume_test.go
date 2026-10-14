package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xaxa/svyaz-cli/internal/mcp"
	"github.com/xaxa/svyaz-cli/internal/session"
)

func TestRenderMCPView(t *testing.T) {
	servers := []mcp.MCPServerInfo{
		{
			Name:       "Neon",
			Status:     "✓",
			AuthNeeded: true,
			Tools:      []string{"run_sql", "run_sql_transaction", "describe_table_schema", "get_database_tables", "prepare_database_migration", "complete_database_migration"},
		},
		{
			Name:       "codebase-memory-mcp",
			Status:     "✓",
			AuthNeeded: false,
			Tools:      []string{"index_repository", "search_graph", "query_graph", "trace_path", "get_code_snippet", "get_graph_schema"},
		},
	}

	view := RenderMCPView(servers, "~/.gemini/config/plugins", 0, "", 90)

	if !strings.Contains(view, "MCP Servers") {
		t.Errorf("expected title 'MCP Servers'")
	}
	if !strings.Contains(view, "Plugins (~/.gemini/config/plugins)") {
		t.Errorf("expected location label")
	}
	if !strings.Contains(view, "[Auth Needed]") {
		t.Errorf("expected '[Auth Needed]' for Neon")
	}
	if !strings.Contains(view, "+1 more") {
		t.Errorf("expected '+1 more' tool count for Neon")
	}
	if !strings.Contains(view, "Keyboard: ↑/↓ Navigate  enter Actions  esc Exit") {
		t.Errorf("expected keyboard hints")
	}
}

func TestRenderResumeView(t *testing.T) {
	sessions := []session.Session{
		{
			ID:        "abc12345",
			Title:     "Add interactive MCP view",
			UpdatedAt: time.Now(),
			Messages: []session.SessionMessage{
				{Role: "user", Content: "Add interactive MCP view"},
				{Role: "agent", Content: "Done"},
			},
		},
	}

	view := RenderResumeView(sessions, 0, 90)

	if !strings.Contains(view, "Chat Sessions") {
		t.Errorf("expected 'Chat Sessions'")
	}
	if !strings.Contains(view, "Add interactive MCP view") {
		t.Errorf("expected session title")
	}
	if !strings.Contains(view, "[2 msgs]") {
		t.Errorf("expected '[2 msgs]' count")
	}
	if !strings.Contains(view, "Keyboard: ↑/↓ Navigate  enter Resume  d Delete  esc Exit") {
		t.Errorf("expected keyboard hints")
	}
}

func TestTUI_MCPAndResumeNavigation(t *testing.T) {
	info := HeaderInfo{
		CurrentModel: "deepseek-coder",
		WorkDir:      "/test/dir",
		ModelStatus:  "Ready",
	}
	m := NewModel(info)
	m.width = 80
	m.height = 24
	m.ready = true

	// 1. Enter /mcp command
	m.input.SetValue("/mcp")
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.state != StateMCPMenu {
		t.Errorf("expected StateMCPMenu, got %v", m.state)
	}
	if len(m.mcpServers) == 0 {
		t.Errorf("expected discovered mcpServers, got 0")
	}

	// 2. Esc out of MCP menu
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(Model)
	if m.state != StateIdle {
		t.Errorf("expected StateIdle after Esc, got %v", m.state)
	}

	// 3. Enter /resume command
	m.input.SetValue("/resume")
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)
	if m.state != StateResumeMenu {
		t.Errorf("expected StateResumeMenu, got %v", m.state)
	}

	// 4. Esc out of Resume menu
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(Model)
	if m.state != StateIdle {
		t.Errorf("expected StateIdle after Esc, got %v", m.state)
	}
}
