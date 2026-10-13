package llm

import "encoding/json"

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// FunctionCall represents the function invoked by the model
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string representation
}

// ToolCall represents a tool invocation in OpenAI format
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolFunctionSpec describes a tool/function definition
type ToolFunctionSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// ToolDefinition defines a callable tool in OpenAI format
type ToolDefinition struct {
	Type     string           `json:"type"` // "function"
	Function ToolFunctionSpec `json:"function"`
}

// ChatCompletionRequest is the payload sent to /v1/chat/completions
type ChatCompletionRequest struct {
	Model       string           `json:"model"`
	Messages    []ChatMessage    `json:"messages"`
	Stream      bool             `json:"stream,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
	MaxTokens   *int             `json:"max_tokens,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	ToolChoice  any              `json:"tool_choice,omitempty"`
}

// ChatUsage holds token usage numbers
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatChoice is one generated choice
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatCompletionResponse is the response from non-streaming /v1/chat/completions
type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
}

// ChatChunkDelta represents token deltas in SSE stream
type ChatChunkDelta struct {
	Role      Role       `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ChatChunkChoice represents chunk choices in streaming
type ChatChunkChoice struct {
	Index        int            `json:"index"`
	Delta        ChatChunkDelta `json:"delta"`
	FinishReason *string        `json:"finish_reason"`
}

// ChatCompletionChunk represents a single chunk in SSE stream
type ChatCompletionChunk struct {
	ID      string            `json:"id"`
	Model   string            `json:"model"`
	Choices []ChatChunkChoice `json:"choices"`
}

// ModelMetadata represents a model discovered from /v1/models
type ModelMetadata struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	ContextWindow       int      `json:"context_window"`
	Available           bool     `json:"available"`
	SupportedParameters []string `json:"supported_parameters"`
}

// SupportsTools returns true if the model declares support for tool calling
func (m *ModelMetadata) SupportsTools() bool {
	for _, p := range m.SupportedParameters {
		if p == "tools" {
			return true
		}
	}
	return false
}

// ModelsListResponse is returned by /v1/models
type ModelsListResponse struct {
	Data []ModelMetadata `json:"data"`
}

// RawJSON helper
func ToJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
