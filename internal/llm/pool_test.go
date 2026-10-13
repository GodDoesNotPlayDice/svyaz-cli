package llm

import (
	"testing"
	"time"
)

func TestModelPool_BasicOrder(t *testing.T) {
	primary := "codestral"
	fallbacks := []string{"devstral-2", "gemini-3.8-flash"}
	pool := NewModelPool(primary, fallbacks, 1*time.Second)

	model, err := pool.GetNextAvailableModel(false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "codestral" {
		t.Errorf("expected primary model 'codestral', got '%s'", model)
	}
}

func TestModelPool_CircuitBreakerOn429(t *testing.T) {
	primary := "codestral"
	fallbacks := []string{"gemini-3.8-flash", "command-r-2"}
	cooldown := 500 * time.Millisecond
	pool := NewModelPool(primary, fallbacks, cooldown)

	// Report 429 on primary
	pool.ReportFailure("codestral", 429)

	if !pool.IsOnCooldown("codestral") {
		t.Errorf("expected codestral to be on cooldown")
	}

	// Next should be gemini-3.8-flash
	next, err := pool.GetNextAvailableModel(false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "gemini-3.8-flash" {
		t.Errorf("expected 'gemini-3.8-flash', got '%s'", next)
	}

	// Wait for cooldown to expire
	time.Sleep(600 * time.Millisecond)

	if pool.IsOnCooldown("codestral") {
		t.Errorf("expected codestral cooldown to have expired")
	}

	recheck, err := pool.GetNextAvailableModel(false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recheck != "codestral" {
		t.Errorf("expected primary 'codestral' to regain priority after cooldown, got '%s'", recheck)
	}
}

func TestModelPool_FilterRequiresTools(t *testing.T) {
	pool := NewModelPool("tiny-model-no-tools", []string{"tool-model"}, 1*time.Second)

	pool.UpdateAvailableModels([]ModelMetadata{
		{
			ID:                  "tiny-model-no-tools",
			Available:           true,
			SupportedParameters: []string{"stream", "temperature"},
		},
		{
			ID:                  "tool-model",
			Available:           true,
			SupportedParameters: []string{"stream", "tools"},
		},
	})

	// When requiring tools, it should skip tiny-model-no-tools
	model, err := pool.GetNextAvailableModel(true, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "tool-model" {
		t.Errorf("expected 'tool-model' because it supports tools, got '%s'", model)
	}
}
