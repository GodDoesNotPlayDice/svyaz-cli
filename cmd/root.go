package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xaxa/svyaz-cli/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "svyaz",
	Short: "svyaz (Связь) - Lightweight, high-performance agentic CLI optimized for FreeLLMAPI",
	Long: `svyaz-cli (Связь) is an interactive agentic development assistant built in Go.
Integrates a 4-layer deterministic harness, 3 sub-agent architecture (Planner, Coder, Reviewer),
and extreme token economy with Token Savior, Graphify, and Engram.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// By default, running 'svyaz' opens the interactive chat TUI
		return chatCmd.RunE(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "configuration file (default ~/.svyaz/config.yaml)")
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load configuration: %v\n", err)
	}
}
