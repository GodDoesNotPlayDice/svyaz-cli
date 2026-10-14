package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveLLMEnv(t *testing.T) {
	// Backup env
	origKey := os.Getenv("FREE_LLM_API_KEY")
	origBase := os.Getenv("FREE_LLM_BASE_URL")
	origServer := os.Getenv("FREE_LLM_SERVER_URL")
	defer func() {
		os.Setenv("FREE_LLM_API_KEY", origKey)
		os.Setenv("FREE_LLM_BASE_URL", origBase)
		os.Setenv("FREE_LLM_SERVER_URL", origServer)
	}()

	t.Run("defaults when empty", func(t *testing.T) {
		os.Unsetenv("FREE_LLM_API_KEY")
		os.Unsetenv("API_KEY")
		os.Unsetenv("FREE_LLM_BASE_URL")
		os.Unsetenv("FREE_LLM_SERVER_URL")
		os.Unsetenv("SERVER_URL")

		key, base := resolveLLMEnv()
		if key == "" {
			t.Errorf("expected default key, got empty")
		}
		if base != "http://147.224.225.246:3001/v1" {
			t.Errorf("expected default base URL, got %s", base)
		}
	})

	t.Run("server_url without v1 automatically appends v1", func(t *testing.T) {
		os.Unsetenv("FREE_LLM_BASE_URL")
		os.Setenv("FREE_LLM_SERVER_URL", "http://example.com:3001")

		_, base := resolveLLMEnv()
		if base != "http://example.com:3001/v1" {
			t.Errorf("expected http://example.com:3001/v1, got %s", base)
		}
	})

	t.Run("custom base_url and api_key", func(t *testing.T) {
		os.Setenv("FREE_LLM_API_KEY", "test-key-123")
		os.Setenv("FREE_LLM_BASE_URL", "http://custom-url/v1")

		cfg := DefaultConfig()
		if cfg.LLM.APIKey != "test-key-123" {
			t.Errorf("expected test-key-123, got %s", cfg.LLM.APIKey)
		}
		if cfg.LLM.BaseURL != "http://custom-url/v1" {
			t.Errorf("expected http://custom-url/v1, got %s", cfg.LLM.BaseURL)
		}
	})
}

func TestLoad_WithYAMLAndEnv(t *testing.T) {
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "config.yaml")

	yamlContent := `
llm:
  provider: "freellmapi"
  base_url: "${FREE_LLM_BASE_URL}"
  api_key: "${FREE_LLM_API_KEY}"
  primary_model: "test-model"
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("writing temp yaml: %v", err)
	}

	os.Setenv("FREE_LLM_BASE_URL", "http://147.224.225.246:3001/v1")
	os.Setenv("FREE_LLM_API_KEY", "freellmapi-d1a9aacfeeeec76524e36844e64b975ee9f30934ba79b5ba")

	cfg, err := Load(yamlFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.LLM.BaseURL != "http://147.224.225.246:3001/v1" {
		t.Errorf("expected base_url to be expanded, got %s", cfg.LLM.BaseURL)
	}
	if cfg.LLM.APIKey != "freellmapi-d1a9aacfeeeec76524e36844e64b975ee9f30934ba79b5ba" {
		t.Errorf("expected api_key to be expanded, got %s", cfg.LLM.APIKey)
	}
	if cfg.LLM.PrimaryModel != "test-model" {
		t.Errorf("expected primary_model test-model, got %s", cfg.LLM.PrimaryModel)
	}
}
