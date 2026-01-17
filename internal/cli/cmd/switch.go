package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

var switchCmd = &cobra.Command{
	Use:   "switch <profile-name> [claude-flags...]",
	Short: "Switch to a different profile (internal use)",
	Long: `Switches the active profile and executes Claude.
This command is intended for internal use and is hidden from the help menu.`,
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		claudeFlags := args[1:]

		// noRun is a persistent flag on the root command, so its value
		// is populated in the global 'noRun' variable.
		return cli.HandleSwitch(profileName, claudeFlags, noRun)
	},
	// Allow passthrough flags for claude
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
