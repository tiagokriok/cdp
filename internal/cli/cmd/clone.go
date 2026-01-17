package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/ui"
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
  cdp clone work work-backup
  cdp clone personal personal-test`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		dest := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
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
