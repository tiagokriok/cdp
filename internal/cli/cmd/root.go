package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/cli"
	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/ui"
)

var noRun bool

var rootCmd = &cobra.Command{
	Use:   "hat",
	Short: "A CLI tool to manage multiple Claude Code profiles",
	Long: `hat, Claude Code profile switcher is a Go CLI tool that manages multiple Claude Code profiles,
enabling seamless switching between different configurations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		migrated, err := config.MigrateLegacyConfigDir()
		if err != nil {
			return err
		}
		if migrated {
			ui.Info("Moved ~/.cdp to ~/.hat")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cli.RunInteractiveMenu(noRun)
		}
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noRun, "no-run", false, "Switch profile without running Claude")
}

// GetRootCmd returns the root command for testing purposes
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// ResetRootCmd resets the root command state for testing
func ResetRootCmd() {
	noRun = false
}
