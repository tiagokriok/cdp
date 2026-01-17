package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
)

var (
	templateFlag    string
	importFromFlag  string
	descriptionFlag string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <profile-name> [description]",
	Short: "Create a new profile",
	Long: `Creates a new profile directory with a default configuration.

Create from scratch with optional template:
  cdp create work "Work profile" --template restrictive
  cdp create personal "Personal" --template permissive

Import existing Claude configuration:
  cdp create work --import-from ~/.config/claude-code --description "Work profile"
  cdp create work --import-from ~/backups/claude-2024

Available templates: restrictive, permissive`,
	Args: cobra.RangeArgs(1, 2), // Expect 1 or 2 arguments
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		description := descriptionFlag

		// Description priority: flag > positional arg
		if description == "" && len(args) > 1 {
			description = args[1]
		}

		// Validate flag compatibility
		if importFromFlag != "" && templateFlag != "" {
			return fmt.Errorf("cannot use --import-from with --template")
		}

		// Route to appropriate handler
		if importFromFlag != "" {
			return cli.HandleImport(importFromFlag, profileName, description)
		}

		return cli.HandleCreateWithTemplate(profileName, description, templateFlag)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Template to apply (restrictive, permissive)")
	createCmd.Flags().StringVar(&importFromFlag, "import-from", "", "Import existing Claude config from directory")
	createCmd.Flags().StringVarP(&descriptionFlag, "description", "d", "", "Profile description")
}
