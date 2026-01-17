package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/ui"
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename an existing profile",
	Long: `Renames a profile to a new name.

Note: You cannot rename the currently active profile.
Switch to another profile first.

Example:
  cdp rename old-work new-work`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
		}

		pm := config.NewProfileManager(cfg)

		if err := pm.RenameProfile(oldName, newName); err != nil {
			return fmt.Errorf("failed to rename profile: %w", err)
		}

		ui.Success(fmt.Sprintf("Profile '%s' renamed to '%s'", oldName, newName))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
