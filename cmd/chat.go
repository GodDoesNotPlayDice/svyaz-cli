package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/xaxa/svyaz-cli/internal/ui"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Inicia la interfaz interactiva TUI de svyaz",
	RunE: func(cmd *cobra.Command, args []string) error {
		headerInfo := ui.HeaderInfo{
			Version:      "v0.1.0",
			RepoName:     ui.GetCurrentDirName(),
			GitBranch:    ui.GetCurrentGitBranch(),
			CurrentModel: "deepseek-coder",
			TokensUsed:   1200,
			TokenLimit:   32000,
		}

		if cfg != nil && cfg.LLM.Model != "" {
			headerInfo.CurrentModel = cfg.LLM.Model
		}

		model := ui.NewModel(headerInfo)
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error ejecutando TUI: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
