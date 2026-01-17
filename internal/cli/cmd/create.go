package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <profile-name> [description]",
	Short: "Create a new profile",
	Long:  `Creates a new profile directory with a default configuration.`,
	Args:  cobra.RangeArgs(1, 2), // Expect 1 or 2 arguments
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		description := ""
		if len(args) > 1 {
			description = args[1]
		}
		return cli.HandleCreate(profileName, description)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
