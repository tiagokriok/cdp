package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/cli"
	"github.com/tiagokriok/hat/internal/prompt"
)

var switchCmd = &cobra.Command{
	Use:   "switch <profile-name> [claude-flags...]",
	Short: "Switch to a different profile (internal use)",
	Long: `Switches the active profile and executes Claude.
This command is intended for internal use and is hidden from the help menu.`,
	Hidden:             true,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		filteredArgs, noRunFlag, yesFlag := extractHatFlags(args)
		prompt.SetAssumeYes(yesFlag)

		if len(filteredArgs) == 0 {
			return fmt.Errorf("profile name is required")
		}

		profileName := filteredArgs[0]
		claudeFlags := filteredArgs[1:]
		return cli.HandleSwitch(profileName, claudeFlags, noRunFlag)
	},
}

func extractHatFlags(args []string) (filtered []string, noRun, yes bool) {
	for _, arg := range args {
		switch arg {
		case "--no-run":
			noRun = true
		case "--yes", "-y":
			yes = true
		default:
			filtered = append(filtered, arg)
		}
	}
	return filtered, noRun, yes
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
