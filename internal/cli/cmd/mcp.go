package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/cli"
)

var mcpAll bool

var mcpCmd = &cobra.Command{
	Use:   "mcp [profile]",
	Short: "List MCP servers of a profile (read-only, secrets never printed)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := ""
		if len(args) == 1 {
			profile = args[0]
		}
		return cli.HandleMCP(profile, mcpAll)
	},
}

var mcpDiffCmd = &cobra.Command{
	Use:   "diff <profile-a> <profile-b>",
	Short: "Compare user-scope MCP servers between two profiles",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleMCPDiff(args[0], args[1])
	},
}

func init() {
	mcpCmd.Flags().BoolVar(&mcpAll, "all", false, "Also list per-project (local scope) servers")
	mcpCmd.AddCommand(mcpDiffCmd)
	rootCmd.AddCommand(mcpCmd)
}
