package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
	"github.com/tiagokriok/cdp/internal/config"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info [profile-name]",
	Short: "Show detailed information about a profile",
	Long: `Displays detailed information about a specific profile.
If no profile name is provided, it shows information for the current active profile.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := ""
		if len(args) > 0 {
			profileName = args[0]
		} else {
			cfg, err := config.Load()
			if err != nil {
				// config.Load() already provides a good error if not initialized
				return err
			}
			profileName = cfg.GetCurrentProfile()
			if profileName == "" {
				return fmt.Errorf("no profile name specified and no profile is currently active")
			}
		}
		return cli.HandleInfo(profileName)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
