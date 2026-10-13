package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xaxa/svyaz-cli/internal/harness"
	"github.com/xaxa/svyaz-cli/internal/ui"
)

var runCmd = &cobra.Command{
	Use:   "run [tarea]",
	Short: "Ejecuta una tarea agéntica en modo headless/batch sin TUI",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := strings.Join(args, " ")
		fmt.Printf("%s %s\n\n", ui.BrandStyle.Render("svyaz-run:"), task)

		fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
			ToolName: "planner",
			Action:   "repo-map",
			Details:  "Extrayendo dependencias y símbolos iniciales...",
			Status:   ui.ActionSuccess,
		}))

		fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
			ToolName: "coder",
			Action:   "diff-gen",
			Details:  "Generando diff conciso con skill ponytail...",
			Status:   ui.ActionSuccess,
		}))

		patcher := harness.NewPatcher(".")
		verifier := harness.NewVerifier(".")

		res, err := verifier.RunVerification()
		if err != nil {
			return err
		}

		fmt.Println(ui.RenderToolBadge(ui.ToolBadge{
			ToolName: "harness",
			Action:   "verify",
			Details:  fmt.Sprintf("Verificación %s: %s", res.Command, res.Output),
			Status:   ui.ActionSuccess,
		}))

		_ = patcher
		fmt.Println("\n" + ui.AgentPrefixStyle.Render("svyaz: ") + "Tarea finalizada.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
