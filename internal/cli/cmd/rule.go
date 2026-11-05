package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/cli"
)

var ruleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Manage directory rules that auto-select a profile",
	Long:  `Directory rules map a path to a profile. Running bare 'hat' inside a mapped directory uses that profile directly.`,
}

var ruleAddCmd = &cobra.Command{
	Use:   "add <path> <profile>",
	Short: "Map a directory to a profile ('.' means the current directory)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleRuleAdd(args[0], args[1])
	},
}

var ruleListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List rules, marking the one matching the current directory",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleRuleList()
	},
}

var ruleRmCmd = &cobra.Command{
	Use:     "rm <path>",
	Short:   "Remove the rule for a directory",
	Aliases: []string{"remove"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.HandleRuleRemove(args[0])
	},
}

func init() {
	ruleCmd.AddCommand(ruleAddCmd, ruleListCmd, ruleRmCmd)
	rootCmd.AddCommand(ruleCmd)
}
