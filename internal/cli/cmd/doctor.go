package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/cli"
)

var doctorCmd = &cobra.Command{
	Use:           "doctor",
	Short:         "Check claude, profiles, logins and rules for problems",
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleDoctor()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
