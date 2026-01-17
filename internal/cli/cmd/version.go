package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time
var Version string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of cdp",
	Long:  `All software has versions. This is cdp's.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Bypassing ui.Header to avoid lipgloss styling issues in tests
		fmt.Println("CDP (Claude Profile Switcher)")
		fmt.Printf("Version: %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

