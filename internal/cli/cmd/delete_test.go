package cmd_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli" // Import internal/cli to access HandleDelete
	"github.com/tiagokriok/cdp/internal/config"
)

func TestDeleteCmd(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Mock stdin for HandleDelete confirmation
	defer mockStdin("y\n")()

	// Initialize config
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Setup a dummy command for testing RunE
	deleteTestCmd := &cobra.Command{
		Use:  "delete",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			return cli.HandleDelete(profileName)
		},
	}

	// Test case 1: Delete an existing profile
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	pm.CreateProfile("test-delete", "Profile to delete")

	err = deleteTestCmd.RunE(deleteTestCmd, []string{"test-delete"})
	if err != nil {
		t.Errorf("Delete existing profile failed: %v", err)
	}
	if pm.ProfileExists("test-delete") {
		t.Error("Profile 'test-delete' should have been deleted")
	}

	// Test case 2: Try to delete a non-existent profile
	err = deleteTestCmd.RunE(deleteTestCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("Deleting a non-existent profile should return an error")
	}

	// Test case 3: Try to delete a profile that is current (should fail from HandleDelete)
	pm.CreateProfile("current-profile", "Current profile")
	cfg.SetCurrentProfile("current-profile")
	defer mockStdin("y\n")()
	err = deleteTestCmd.RunE(deleteTestCmd, []string{"current-profile"})
	if err == nil {
		t.Error("Deleting the current profile should return an error")
	}
}

func TestDeleteCmd_UserConfirms(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize config
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Create a profile
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	pm.CreateProfile("to-delete", "Will be deleted")

	// Mock user confirming deletion with "y"
	defer mockStdin("y\n")()

	// Delete the profile
	err = pm.DeleteProfile("to-delete")
	if err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	if pm.ProfileExists("to-delete") {
		t.Error("Profile should have been deleted")
	}
}

func TestDeleteCmd_InvalidProfileName(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize config
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Try to delete with invalid name
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	err = pm.DeleteProfile("invalid@name")
	if err == nil {
		t.Error("Should fail with invalid profile name")
	}
}
