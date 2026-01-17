package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete <profile-name>",
	Short:   "Delete a profile",
	Long:    `Deletes a profile directory and its contents.`,
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1), // Expect exactly 1 argument
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		return cli.HandleDelete(profileName)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
