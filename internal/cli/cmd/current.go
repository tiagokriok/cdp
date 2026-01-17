package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

// currentCmd represents the current command
var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active profile",
	Long:  `Displays the name of the profile that is currently active.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleCurrent()
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
