package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/executor"
	"github.com/tiagokriok/cdp/internal/ui"
)

// HandleInit initializes the CDP configuration
func HandleInit() error {
	if config.Exists() {
		ui.Info("CDP is already initialized.")
		fmt.Println("Configuration directory:", getConfigDirOrEmpty())
		fmt.Println("Profiles directory:", getProfilesDirOrEmpty())
		return nil
	}

	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to initialize CDP: %w", err)
	}

	ui.Success("CDP initialized successfully!")
	fmt.Println("Configuration directory:", getConfigDirOrEmpty())
	fmt.Println("Profiles directory:", getProfilesDirOrEmpty())
	fmt.Println("\nGet started by creating a profile:")
	fmt.Println("  cdp create <name> [description]")

	return nil
}

// HandleCreate creates a new profile
func HandleCreate(name, description string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)

	if err := pm.CreateProfile(name, description); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	ui.Success(fmt.Sprintf("Profile '%s' created successfully!", name))
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}
	fmt.Printf("Location: %s\n", cfg.GetProfilesDir()+"/"+name)
	fmt.Println("\nSwitch to this profile:")
	fmt.Printf("  cdp %s\n", name)

	return nil
}

// HandleList lists all profiles
func HandleList() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)
	profiles, err := pm.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	currentProfile := cfg.GetCurrentProfile()
	ui.PrintProfileList(profiles, currentProfile)

	return nil
}

// HandleDelete deletes a profile
func HandleDelete(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)

	// Check if profile exists
	if !pm.ProfileExists(name) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete profile '%s'? [y/N]: ", name)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		ui.Info("Deletion cancelled.")
		return nil
	}

	if err := pm.DeleteProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	ui.Success(fmt.Sprintf("Profile '%s' deleted successfully.", name))
	return nil
}

// HandleCurrent shows the current active profile
func HandleCurrent() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	currentProfile := cfg.GetCurrentProfile()

	if currentProfile == "" {
		ui.Info("No profile is currently active.")
		fmt.Println("\nSwitch to a profile:")
		fmt.Println("  cdp <profile-name>")
		return nil
	}

	// Get profile details
	pm := config.NewProfileManager(cfg)
	profile, err := pm.GetProfile(currentProfile)
	if err != nil {
		// Profile might have been deleted
		ui.Warn("Profile directory not found")
		return nil
	}

	ui.PrintProfileInfo(profile, true)

	return nil
}

// HandleInfo shows detailed information about a specific profile
func HandleInfo(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)
	profile, err := pm.GetProfile(name)
	if err != nil {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	currentProfile := cfg.GetCurrentProfile()
	isCurrent := profile.Name == currentProfile

	ui.PrintProfileInfo(profile, isCurrent)

	return nil
}

// HandleSwitch switches to a profile and optionally runs Claude
func HandleSwitch(name string, claudeFlags []string, noRun bool) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)

	// Check if profile exists
	profile, err := pm.GetProfile(name)
	if err != nil {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Validate profile structure
	if err := pm.ValidateProfile(profile); err != nil {
		return fmt.Errorf("profile '%s' is corrupted: %w", name, err)
	}

	// Update current profile by setting field and saving explicitly
	cfg.CurrentProfile = name
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config with new profile: %w", err)
	}

	// Update last used timestamp
	if err := pm.UpdateLastUsed(name); err != nil {
		// Non-fatal error, just log it
		ui.Warn(fmt.Sprintf("Failed to update last used timestamp: %v", err))
	}

	ui.Success(fmt.Sprintf("Switched to profile: %s", name))

	if noRun {
		ui.Info("Use 'claude' to start Claude Code with this profile.")
		return nil
	}

	// Run Claude Code
	ui.Info("Starting Claude Code...")
	exec := executor.NewExecutor()
	return exec.Run(profile.Path, claudeFlags)
}

// Helper functions

func loadConfig() (*config.Config, error) {
	if !config.Exists() {
		return nil, fmt.Errorf("CDP is not initialized. Run 'cdp init' first")
	}
	return config.Load()
}

func getConfigDirOrEmpty() string {
	dir, err := config.GetConfigDir()
	if err != nil {
		return ""
	}
	return dir
}

func getProfilesDirOrEmpty() string {
	dir, err := config.GetDefaultProfilesDir()
	if err != nil {
		return ""
	}
	return dir
}
