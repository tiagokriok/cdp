package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/backup"
	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/ui"
)

var overwriteFlag bool

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore profiles",
	Long: `Manage profile backups.

Commands:
  cdp backup <profile>       - Create a backup of a profile
  cdp backup list            - List all backups
  cdp backup restore <file>  - Restore a profile from backup
  cdp backup delete <file>   - Delete a backup file`,
}

// backupCreateCmd creates a backup
var backupCreateCmd = &cobra.Command{
	Use:   "create <profile>",
	Short: "Create a backup of a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
		}

		bm, err := backup.NewBackupManager(cfg.GetProfilesDir())
		if err != nil {
			return fmt.Errorf("failed to initialize backup manager: %w", err)
		}

		backupPath, err := bm.Backup(profileName)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		ui.Success(fmt.Sprintf("Backup created: %s", backupPath))
		return nil
	},
}

// backupListCmd lists all backups
var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
		}

		bm, err := backup.NewBackupManager(cfg.GetProfilesDir())
		if err != nil {
			return fmt.Errorf("failed to initialize backup manager: %w", err)
		}

		backups, err := bm.List()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			ui.Info("No backups found.")
			fmt.Println("\nCreate a backup with:")
			fmt.Println("  cdp backup create <profile>")
			return nil
		}

		ui.Header(fmt.Sprintf("Found %d backup(s):", len(backups)))
		fmt.Println()

		for _, b := range backups {
			fmt.Printf("  %s\n", ui.ProfileStyle.Render(b.Name))
			fmt.Printf("    Profile: %s\n", b.ProfileName)
			fmt.Printf("    Created: %s\n", b.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("    Size: %s\n", formatBytes(b.Size))
			fmt.Println()
		}

		fmt.Printf("Backup directory: %s\n", bm.GetBackupDir())
		return nil
	},
}

// backupRestoreCmd restores from backup
var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore a profile from backup",
	Long: `Restores a profile from a backup file.

The backup file can be specified as just the filename (if in default backup directory)
or as a full path.

Use --overwrite to replace an existing profile with the same name.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupFile := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
		}

		bm, err := backup.NewBackupManager(cfg.GetProfilesDir())
		if err != nil {
			return fmt.Errorf("failed to initialize backup manager: %w", err)
		}

		// If not a path, assume it's in the backup directory
		if backupFile[0] != '/' && backupFile[0] != '.' {
			backupFile = fmt.Sprintf("%s/%s", bm.GetBackupDir(), backupFile)
		}

		profileName, err := bm.Restore(backupFile, overwriteFlag)
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		ui.Success(fmt.Sprintf("Profile '%s' restored successfully!", profileName))
		return nil
	},
}

// backupDeleteCmd deletes a backup
var backupDeleteCmd = &cobra.Command{
	Use:   "delete <backup-file>",
	Short: "Delete a backup file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupFile := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
		}

		bm, err := backup.NewBackupManager(cfg.GetProfilesDir())
		if err != nil {
			return fmt.Errorf("failed to initialize backup manager: %w", err)
		}

		if err := bm.Delete(backupFile); err != nil {
			return fmt.Errorf("failed to delete backup: %w", err)
		}

		ui.Success(fmt.Sprintf("Backup '%s' deleted.", backupFile))
		return nil
	},
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupDeleteCmd)

	backupRestoreCmd.Flags().BoolVar(&overwriteFlag, "overwrite", false, "Overwrite existing profile if it exists")
}
