package mcp

import (
	"fmt"
	"strings"
)

type ToolParameterSpec struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type ToolDefinition struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Parameters  map[string]ToolParameterSpec `json:"parameters"`
	Required    []string                     `json:"required"`
}

// Validator ensures LLM tool calls don't hallucinate non-existent parameters
type Validator struct {
	tools map[string]ToolDefinition
}

func NewValidator() *Validator {
	return &Validator{
		tools: make(map[string]ToolDefinition),
	}
}

func (v *Validator) RegisterTool(td ToolDefinition) {
	v.tools[td.Name] = td
}

// Validate checks if the tool call arguments match the declared schema
func (v *Validator) Validate(toolName string, args map[string]any) error {
	tool, exists := v.tools[toolName]
	if !exists {
		return fmt.Errorf("tool '%s' is not registered in active MCP servers", toolName)
	}

	// 1. Check for missing required parameters
	for _, req := range tool.Required {
		if _, found := args[req]; !found {
			return fmt.Errorf("missing required parameter '%s' for tool '%s'", req, toolName)
		}
	}

	// 2. Check for hallucinated / unknown parameters
	var invalidParams []string
	for paramKey := range args {
		if _, allowed := tool.Parameters[paramKey]; !allowed {
			invalidParams = append(invalidParams, paramKey)
		}
	}

	if len(invalidParams) > 0 {
		var validList []string
		for k := range tool.Parameters {
			validList = append(validList, k)
		}
		return fmt.Errorf(
			"hallucinated parameters detected [%s]. Tool '%s' only accepts: [%s]",
			strings.Join(invalidParams, ", "),
			toolName,
			strings.Join(validList, ", "),
		)
	}

	return nil
}
