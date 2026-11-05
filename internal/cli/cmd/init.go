package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/cli"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the hat configuration directory and profiles",
	Long: `Creates the main configuration file at ~/.hat/config.yaml
and the profiles directory at ~/.claude-profiles if they do not exist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
