package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

var noRun bool

var rootCmd = &cobra.Command{
	Use:   "cdp",
	Short: "A CLI tool to manage multiple Claude Code profiles",
	Long: `cdp (Claude Profile Switcher) is a Go CLI tool that manages multiple Claude Code profiles,
enabling seamless switching between different configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// No arguments provided, so run the interactive menu.
			return cli.RunInteractiveMenu()
		}
		// If arguments were provided but not handled by a subcommand,
		// it's an error. The pre-parser in main.go should have
		// converted implicit switches.
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra prints the error message by default.
		// We just need to exit with a non-zero status code.
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
