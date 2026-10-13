package agent

import "time"

// Role defines the message sender or sub-agent type
type Role string

const (
	RoleUser        Role = "user"
	RoleAssistant   Role = "assistant"
	RoleSystem      Role = "system"
	RoleTool        Role = "tool"
	RolePlanner     Role = "planner"
	RoleCoder       Role = "coder"
	RoleReviewer    Role = "reviewer"
	RoleOrchestrator Role = "orchestrator"
)

// ToolCall represents a requested function call with arguments
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolResult represents the output from an executed tool
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// Message represents a single chat or system message in the agent conversation
type Message struct {
	Role       Role         `json:"role"`
	Content    string       `json:"content"`
	ToolCalls  []ToolCall   `json:"tool_calls,omitempty"`
	ToolResult *ToolResult  `json:"tool_result,omitempty"`
	Timestamp  time.Time    `json:"timestamp"`
}

// PlanStep represents a granular task produced by Planner
type PlanStep struct {
	Description string   `json:"description"`
	Symbols     []string `json:"symbols"`
	Files       []string `json:"files"`
}

// Plan represents the output produced by the Planner Agent
type Plan struct {
	Summary string     `json:"summary"`
	Steps   []PlanStep `json:"steps"`
}

// Patch represents a unified diff produced by the Coder Agent
type Patch struct {
	FilePath string `json:"file_path"`
	Diff     string `json:"diff"`
	AddCount int    `json:"add_count"`
	DelCount int    `json:"del_count"`
}
