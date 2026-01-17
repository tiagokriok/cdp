package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available profiles",
	Long:    `Lists all profiles found in the profiles directory.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
