package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xaxa/svyaz-cli/internal/ui"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Inspecciona y verifica los servidores MCP configurados",
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todos los servidores MCP configurados y su estado",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.BrandStyle.Render("Servidores MCP configurados:"))
		if cfg == nil || len(cfg.MCPServers) == 0 {
			fmt.Println("  No hay servidores MCP configurados.")
			return
		}

		for name, srv := range cfg.MCPServers {
			fmt.Printf("  • %s: %s %v\n", ui.HeaderAccentValueStyle.Render(name), srv.Command, srv.Args)
		}
	},
}

func init() {
	mcpCmd.AddCommand(mcpListCmd)
	rootCmd.AddCommand(mcpCmd)
}
