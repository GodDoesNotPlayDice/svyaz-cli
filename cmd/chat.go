package cmd

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/xaxa/svyaz-cli/internal/agent"
	"github.com/xaxa/svyaz-cli/internal/llm"
	"github.com/xaxa/svyaz-cli/internal/ui"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start the interactive TUI interface of svyaz",
	RunE: func(cmd *cobra.Command, args []string) error {
		headerInfo := ui.HeaderInfo{
			Version:      "v1.0.0",
			WorkDir:      ui.GetActiveWorkDir(),
			CurrentModel: "codestral",
			ModelStatus:  "Ready",
		}

		if cfg != nil {
			if cfg.LLM.PrimaryModel != "" {
				headerInfo.CurrentModel = cfg.LLM.PrimaryModel
			} else if cfg.LLM.Model != "" {
				headerInfo.CurrentModel = cfg.LLM.Model
			}
		}

		// Initialize LLM Pool & Orchestrator
		var orch *agent.Orchestrator
		if cfg != nil && cfg.LLM.BaseURL != "" {
			cooldownSec := cfg.LLM.CooldownSeconds
			if cooldownSec <= 0 {
				cooldownSec = 60
			}
			pool := llm.NewModelPool(
				headerInfo.CurrentModel,
				cfg.LLM.FallbackPool,
				time.Duration(cooldownSec)*time.Second,
			)

			timeoutSec := cfg.LLM.TimeoutSeconds
			if timeoutSec <= 0 {
				timeoutSec = 60
			}
			client := llm.NewClient(
				cfg.LLM.BaseURL,
				cfg.LLM.APIKey,
				time.Duration(timeoutSec)*time.Second,
				pool,
			)

			// Background sync of available models from FreeLLMAPI
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, _ = client.FetchModels(ctx)
			}()

			orch = agent.NewOrchestrator(client, headerInfo.WorkDir)
		}

		model := ui.NewModel(headerInfo, orch)
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
