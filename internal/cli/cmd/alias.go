package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/ui"
	"github.com/tiagokriok/cdp/pkg/aliases"
)

// aliasCmd represents the alias command
var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage shell aliases for profiles",
	Long: `Manage shell aliases for quick profile switching.

Commands:
  cdp alias install   - Install aliases to your shell RC file
  cdp alias uninstall - Remove aliases from your shell RC file
  cdp alias list      - List currently installed aliases`,
}

// aliasInstallCmd represents the alias install command
var aliasInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install shell aliases for all profiles",
	Long:  `Installs shell aliases to your shell RC file (.bashrc, .zshrc, or config.fish).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
		}

		pm := config.NewProfileManager(cfg)
		profiles, err := pm.ListProfiles()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(profiles) == 0 {
			ui.Info("No profiles found. Create some profiles first.")
			return nil
		}

		// Get profile names
		profileNames := make([]string, len(profiles))
		for i, p := range profiles {
			profileNames[i] = p.Name
		}

		// Generate aliases
		aliasMap := aliases.GenerateDefaultAliases(profileNames)

		// Install
		am, err := aliases.New()
		if err != nil {
			return fmt.Errorf("failed to detect shell: %w", err)
		}

		if err := am.InstallAliases(aliasMap); err != nil {
			return fmt.Errorf("failed to install aliases: %w", err)
		}

		ui.Success("Shell aliases installed!")
		fmt.Printf("RC file: %s\n", am.GetRCFile())
		fmt.Printf("Shell: %s\n", am.GetShellType())
		fmt.Println("\nInstalled aliases:")
		for profile, shortcut := range aliasMap {
			fmt.Printf("  %s -> cdp %s\n", shortcut, profile)
		}
		fmt.Println("\nRestart your shell or run:")
		fmt.Printf("  source %s\n", am.GetRCFile())

		return nil
	},
}

// aliasUninstallCmd represents the alias uninstall command
var aliasUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove shell aliases",
	Long:  `Removes all cdp aliases from your shell RC file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		am, err := aliases.New()
		if err != nil {
			return fmt.Errorf("failed to detect shell: %w", err)
		}

		if !am.IsInstalled() {
			ui.Info("No cdp aliases are installed.")
			return nil
		}

		if err := am.UninstallAliases(); err != nil {
			return fmt.Errorf("failed to uninstall aliases: %w", err)
		}

		ui.Success("Shell aliases removed!")
		fmt.Printf("RC file: %s\n", am.GetRCFile())
		fmt.Println("\nRestart your shell or run:")
		fmt.Printf("  source %s\n", am.GetRCFile())

		return nil
	},
}

// aliasListCmd represents the alias list command
var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed shell aliases",
	Long:  `Lists all cdp aliases currently installed in your shell RC file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		am, err := aliases.New()
		if err != nil {
			return fmt.Errorf("failed to detect shell: %w", err)
		}

		installedAliases, err := am.ListAliases()
		if err != nil {
			return fmt.Errorf("failed to list aliases: %w", err)
		}

		if len(installedAliases) == 0 {
			ui.Info("No cdp aliases are installed.")
			fmt.Println("\nInstall aliases with:")
			fmt.Println("  cdp alias install")
			return nil
		}

		ui.Header("Installed aliases:")
		fmt.Println()
		for shortcut, profile := range installedAliases {
			fmt.Printf("  %s -> cdp %s\n", shortcut, profile)
		}
		fmt.Printf("\nRC file: %s\n", am.GetRCFile())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(aliasCmd)
	aliasCmd.AddCommand(aliasInstallCmd)
	aliasCmd.AddCommand(aliasUninstallCmd)
	aliasCmd.AddCommand(aliasListCmd)
}
