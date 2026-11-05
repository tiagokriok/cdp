package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/ui"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone <source> <destination>",
	Short: "Clone an existing profile",
	Long: `Creates a copy of an existing profile with a new name.

The cloned profile will have:
- All configuration files from the source
- Reset usage count and last used timestamp
- New creation timestamp

Example:
  hat clone work work-backup
  hat clone personal personal-test`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		dest := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("hat not initialized. Run 'hat init' first")
		}

		pm := config.NewProfileManager(cfg)

		if err := pm.CloneProfile(source, dest); err != nil {
			return fmt.Errorf("failed to clone profile: %w", err)
		}

		ui.Success(fmt.Sprintf("Profile '%s' cloned to '%s'", source, dest))
		fmt.Printf("Location: %s/%s\n", cfg.GetProfilesDir(), dest)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
