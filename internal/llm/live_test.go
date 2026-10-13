package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLive_FreeLLMAPI_Connection(t *testing.T) {
	if os.Getenv("RUN_LIVE_TESTS") != "true" {
		t.Skip("skipping live network test; set RUN_LIVE_TESTS=true to run")
	}

	apiKey := "freellmapi-d1a9aacfeeeec76524e36844e64b975ee9f30934ba79b5ba"
	baseURL := "http://147.224.225.246:3001/v1"

	pool := NewModelPool("codestral", []string{"gemini-3.8-flash", "command-r-2"}, 60*time.Second)
	client := NewClient(baseURL, apiKey, 30*time.Second, pool)

	// Test streaming
	var tokens []string
	resp, err := client.ExecuteWithFallback(context.Background(), ChatCompletionRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Responde únicamente la palabra: 'Conectado'"},
		},
	}, func(token string) {
		tokens = append(tokens, token)
	})

	if err != nil {
		t.Fatalf("live FreeLLMAPI call failed: %v", err)
	}

	if resp == nil || len(resp.Choices) == 0 {
		t.Fatalf("expected valid choices in response")
	}

	fullResponse := strings.Join(tokens, "")
	t.Logf("Modelo usado: %s, Respuesta: %s", resp.Model, fullResponse)

	if !strings.Contains(strings.ToLower(fullResponse), "conectado") {
		t.Errorf("expected response to contain 'conectado', got '%s'", fullResponse)
	}
}
