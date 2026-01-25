package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

var switchCmd = &cobra.Command{
	Use:   "switch <profile-name> [claude-flags...]",
	Short: "Switch to a different profile (internal use)",
	Long: `Switches the active profile and executes Claude.
This command is intended for internal use and is hidden from the help menu.`,
	Hidden:             true,
	DisableFlagParsing: true, // Pass all args through unchanged to support Claude flags
	RunE: func(cmd *cobra.Command, args []string) error {
		// Manually extract --no-run flag since DisableFlagParsing is true
		var filteredArgs []string
		noRunFlag := false
		for _, arg := range args {
			if arg == "--no-run" {
				noRunFlag = true
			} else {
				filteredArgs = append(filteredArgs, arg)
			}
		}

		if len(filteredArgs) == 0 {
			return fmt.Errorf("profile name is required")
		}

		profileName := filteredArgs[0]
		claudeFlags := filteredArgs[1:]
		return cli.HandleSwitch(profileName, claudeFlags, noRunFlag)
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
