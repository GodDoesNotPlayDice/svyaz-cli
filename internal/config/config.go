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
	Model           string   `yaml:"model"`
	FallbackModels  []string `yaml:"fallback_models"`
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
	return &Config{
		LLM: LLMConfig{
			Provider:       "freellmapi",
			BaseURL:        "https://api.freellm.example/v1",
			APIKey:         os.Getenv("FREE_LLM_API_KEY"),
			Model:          "deepseek-coder",
			FallbackModels: []string{"qwen2.5-coder", "llama-3.1-70b"},
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
