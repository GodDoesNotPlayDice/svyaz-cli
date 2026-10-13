package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type LLMConfig struct {
	Provider        string   `yaml:"provider"`
	BaseURL         string   `yaml:"base_url"`
	APIKey          string   `yaml:"api_key"`
	Model           string   `yaml:"model"` // Primary model (legacy / shortcut)
	PrimaryModel    string   `yaml:"primary_model"`
	FallbackPool    []string `yaml:"fallback_pool"`
	CooldownSeconds int      `yaml:"cooldown_seconds"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
	MaxRetries      int      `yaml:"max_retries"`
}

type MCPServerConfig struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env,omitempty"`
}

type SkillsConfig struct {
	Active []string `yaml:"active"`
}

type AgentConfig struct {
	AutoCompactContext  bool `yaml:"auto_compact_context"`
	MaxOutputLines      int  `yaml:"max_output_lines"`
	RequireConfirmation bool `yaml:"require_confirmation"`
}

type Config struct {
	LLM        LLMConfig                  `yaml:"llm"`
	MCPServers map[string]MCPServerConfig `yaml:"mcp_servers"`
	Skills     SkillsConfig               `yaml:"skills"`
	Agent      AgentConfig                `yaml:"agent"`
}

// DefaultConfig returns a sane default configuration optimized for FreeLLMAPI
func DefaultConfig() *Config {
	apiKey := os.Getenv("FREE_LLM_API_KEY")
	if apiKey == "" {
		apiKey = "freellmapi-d1a9aacfeeeec76524e36844e64b975ee9f30934ba79b5ba"
	}

	return &Config{
		LLM: LLMConfig{
			Provider:        "freellmapi",
			BaseURL:         "http://147.224.225.246:3001/v1",
			APIKey:          apiKey,
			Model:           "codestral",
			PrimaryModel:    "codestral",
			CooldownSeconds: 60,
			FallbackPool: []string{
				// Tier 1: Coding & Deep Reasoning
				"codestral",
				"devstral-2",
				"north-mini-code",
				"qwen2.5-coder-32b",
				"gemini-3.8-flash",
				"gemini-3.7-flash",
				"deepseek-r1-distill-qwen-32b",
				"qwq-32b",
				// Tier 2: General High-Capacity with Tools
				"mistral-large-3",
				"mistral-medium-3.5",
				"mistral-small-4",
				"mistral-small-3.1-24b",
				"gemini-3.6-flash",
				"gemini-3.5-flash",
				"gemini-3.5-flash-lite",
				"gemini-3.1-flash-lite",
				"command-a",
				"command-a-2",
				"command-a-reasoning",
				"command-r-2",
				"command-r",
				"qwen3.8-27b",
				"gpt-oss-120b",
				"nemotron-3-super-120b",
				"nemotron-3.5-lightning",
				"glm-4.7-flash",
				"llama-3.3-70b-fp8-fast",
				"llama-4-scout",
				// Tier 3: Fast Lightweight Fallbacks
				"ministral-3-14b",
				"ministral-3-8b",
				"ministral-3-3b",
				"sensenova-6.8-flash-lite",
				"poolside-laguna-xs-2.1",
				"dots3-note-preview",
				"command-r7b",
				"llama-3.1-8b",
				"llama-3.2-3b",
				"gpt-oss-20b",
			},
			TimeoutSeconds: 60,
			MaxRetries:     3,
		},
		MCPServers: map[string]MCPServerConfig{
			"token_savior": {
				Command: "npx",
				Args:    []string{"-y", "token-savior-mcp"},
			},
			"graphify": {
				Command: "graphify-mcp",
				Args:    []string{"--root", "."},
			},
			"engram": {
				Command: "engram-mcp",
				Args:    []string{"--db", "~/.svyaz/memory.db"},
			},
		},
		Skills: SkillsConfig{
			Active: []string{"ponytail", "ast-navigator"},
		},
		Agent: AgentConfig{
			AutoCompactContext:  true,
			MaxOutputLines:      100,
			RequireConfirmation: true,
		},
	}
}

// Load loads config from a specified path or the standard ~/.svyaz/config.yaml location
func Load(customPath string) (*Config, error) {
	cfg := DefaultConfig()

	path := customPath
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			localPath := filepath.Join(".svyaz", "config.yaml")
			homePath := filepath.Join(homeDir, ".svyaz", "config.yaml")
			if _, err := os.Stat(localPath); err == nil {
				path = localPath
			} else if _, err := os.Stat(homePath); err == nil {
				path = homePath
			}
		}
	}

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && customPath == "" {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config from %s: %w", path, err)
	}

	expanded := os.ExpandEnv(string(data))
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parsing config yaml %s: %w", path, err)
	}

	return cfg, nil
}
