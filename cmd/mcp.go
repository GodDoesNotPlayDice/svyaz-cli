package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xaxa/svyaz-cli/internal/mcp"
	"github.com/xaxa/svyaz-cli/internal/ui"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Inspect and verify configured MCP servers and active tools",
	Run: func(cmd *cobra.Command, args []string) {
		servers, location := mcp.DiscoverServers()
		fmt.Println(ui.RenderMCPView(servers, location, 0, "", 100))
	},
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured MCP servers and their status",
	Run: func(cmd *cobra.Command, args []string) {
		servers, location := mcp.DiscoverServers()
		fmt.Println(ui.RenderMCPView(servers, location, 0, "", 100))
	},
}

func init() {
	mcpCmd.AddCommand(mcpListCmd)
	rootCmd.AddCommand(mcpCmd)
}
