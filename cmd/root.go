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
	Short: "svyaz (Связь) - CLI agéntico ligero y de alto rendimiento optimizado para FreeLLMAPI",
	Long: `svyaz-cli (Связь) es un asistente interactivo de desarrollo agéntico compilado en Go.
Integra un arnés determinista de 4 capas, arquitectura de 3 sub-agentes (Planner, Coder, Reviewer)
y optimización extrema de tokens con Token Savior, Graphify y Engram.`,
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "archivo de configuración (default ~/.svyaz/config.yaml)")
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Aviso: no se pudo cargar la configuración: %v\n", err)
	}
}
