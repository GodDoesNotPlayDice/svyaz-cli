package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xaxa/svyaz-cli/internal/agent"
	"github.com/xaxa/svyaz-cli/internal/llm"
	"github.com/xaxa/svyaz-cli/internal/ui"
)

var runCmd = &cobra.Command{
	Use:   "run [tarea]",
	Short: "Ejecuta una tarea agéntica en modo headless/batch sin TUI",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := strings.Join(args, " ")
		fmt.Printf("%s %s\n\n", ui.BrandStyle.Render("svyaz-run:"), task)

		model := "codestral"
		if cfg != nil {
			if cfg.LLM.PrimaryModel != "" {
				model = cfg.LLM.PrimaryModel
			} else if cfg.LLM.Model != "" {
				model = cfg.LLM.Model
			}
		}

		cooldownSec := 60
		timeoutSec := 60
		var fallbackPool []string

		if cfg != nil {
			if cfg.LLM.CooldownSeconds > 0 {
				cooldownSec = cfg.LLM.CooldownSeconds
			}
			if cfg.LLM.TimeoutSeconds > 0 {
				timeoutSec = cfg.LLM.TimeoutSeconds
			}
			fallbackPool = cfg.LLM.FallbackPool
		}

		baseURL := "http://147.224.225.246:3001/v1"
		apiKey := "freellmapi-d1a9aacfeeeec76524e36844e64b975ee9f30934ba79b5ba"
		if cfg != nil {
			if cfg.LLM.BaseURL != "" {
				baseURL = cfg.LLM.BaseURL
			}
			if cfg.LLM.APIKey != "" {
				apiKey = cfg.LLM.APIKey
			}
		}

		pool := llm.NewModelPool(model, fallbackPool, time.Duration(cooldownSec)*time.Second)
		client := llm.NewClient(baseURL, apiKey, time.Duration(timeoutSec)*time.Second, pool)

		// Sync available models
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, _ = client.FetchModels(ctx)
		cancel()

		orch := agent.NewOrchestrator(client, ".")
		orch.Callbacks = agent.OrchestratorCallbacks{
			OnStatusChange: func(status string) {
				fmt.Printf("• %s\n", status)
			},
			OnPlanReady: func(plan *agent.Plan) {
				fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
					ToolName: "planner",
					Action:   "repo-map",
					Details:  plan.Summary,
					Status:   ui.ActionSuccess,
				}))
			},
			OnPatchReady: func(patch *agent.Patch) {
				fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
					ToolName: "git",
					Action:   "dry-run",
					Details:  fmt.Sprintf("Parche verificado para %s (+%d/-%d)", patch.FilePath, patch.AddCount, patch.DelCount),
					Status:   ui.ActionSuccess,
				}))
			},
			OnFallback: func(ev llm.FallbackEvent) {
				fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
					ToolName: "fallback",
					Action:   "circuit-breaker",
					Details:  fmt.Sprintf("%s (429) -> %s", ev.FailedModel, ev.NextModel),
					Status:   ui.ActionWarning,
				}))
			},
		}

		// Run pipeline with intent detection
		runCtx := context.Background()
		execRes, err := orch.ProcessInput(runCtx, task, nil)
		if err != nil {
			return fmt.Errorf("error procesando tarea: %w", err)
		}

		if execRes.Intent == agent.IntentChat {
			fmt.Printf("\n%s\n\n", execRes.Response)
			fmt.Println(ui.AgentPrefixStyle.Render("svyaz: ") + "Listo.")
			return nil
		}

		patch := execRes.Patch
		if patch == nil {
			return fmt.Errorf("no se generó parche para la tarea")
		}

		fmt.Printf("\n%s\n%s\n", ui.MetaTitleStyle.Render("=== DIFF GENERADO ==="), patch.Diff)

		// Apply and verify
		res, err := orch.ApplyAndVerify(patch)
		if err != nil {
			return fmt.Errorf("error aplicando o verificando parche: %w", err)
		}

		if !res.Success {
			fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
				ToolName: "verifier",
				Action:   "check",
				Details:  fmt.Sprintf("%s falló: %s", res.Command, res.Output),
				Status:   ui.ActionError,
			}))
			return fmt.Errorf("verificación determinista falló")
		}

		fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
			ToolName: "verifier",
			Action:   "check",
			Details:  fmt.Sprintf("%s: Verificación exitosa", res.Command),
			Status:   ui.ActionSuccess,
		}))

		fmt.Println("\n" + ui.AgentPrefixStyle.Render("svyaz: ") + "Tarea finalizada con éxito.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
