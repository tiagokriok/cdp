package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/ui"
)

// templatesCmd represents the templates command
var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "List available profile templates",
	Long:  `Lists all built-in and custom profile templates that can be used with 'hat create --template'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm := config.NewTemplateManager()
		templates, err := tm.ListTemplates()
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}

		if len(templates) == 0 {
			ui.Info("No templates available.")
			return nil
		}

		ui.Header("Available templates:")
		fmt.Println()
		for _, name := range templates {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println()
		fmt.Println("Use with: hat create <name> --template <template-name>")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
}
