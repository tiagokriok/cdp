package cli

import (
	"errors"
	"fmt"
	"github.com/tiagokriok/hat/internal/prompt"
	"os"
	"strings"

	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/executor"
	"github.com/tiagokriok/hat/internal/ui"
)

var ErrNotInitialized = errors.New("hat is not initialized. Run 'hat init' first")

// HandleInit initializes the hat configuration
func HandleInit() error {
	if config.Exists() {
		ui.Info("hat is already initialized.")
		fmt.Println("Configuration directory:", getConfigDirOrEmpty())
		fmt.Println("Profiles directory:", getProfilesDirOrEmpty())
		return nil
	}

	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to initialize hat: %w", err)
	}

	ui.Success("hat initialized successfully!")
	fmt.Println("Configuration directory:", getConfigDirOrEmpty())
	fmt.Println("Profiles directory:", getProfilesDirOrEmpty())
	fmt.Println("\nGet started by creating a profile:")
	fmt.Println("  hat create <name> [description]")

	return nil
}

// HandleCreate creates a new profile
func HandleCreate(name, description string) error {
	return HandleCreateWithTemplate(name, description, "")
}

// HandleCreateWithTemplate creates a new profile with an optional template
func HandleCreateWithTemplate(name, description, template string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)

	if err := pm.CreateProfileWithTemplate(name, description, template); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	ui.Success(fmt.Sprintf("Profile '%s' created successfully!", name))
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}
	if template != "" {
		fmt.Printf("Template: %s\n", template)
	}
	fmt.Printf("Location: %s\n", cfg.GetProfilesDir()+"/"+name)
	fmt.Println("\nSwitch to this profile:")
	fmt.Printf("  hat %s\n", name)

	return nil
}

// HandleImport imports an existing Claude configuration into a new profile
func HandleImport(sourcePath, name, description string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)

	if err := pm.ImportProfile(sourcePath, name, description); err != nil {
		return fmt.Errorf("failed to import profile: %w", err)
	}

	ui.Success(fmt.Sprintf("Profile '%s' imported successfully!", name))
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}
	fmt.Printf("Location: %s\n", cfg.GetProfilesDir()+"/"+name)
	fmt.Println("\nSwitch to this profile:")
	fmt.Printf("  hat %s\n", name)

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

	if !pm.ProfileExists(name) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	if !prompt.Confirm(fmt.Sprintf("Are you sure you want to delete profile '%s'?", name), false) {
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
		fmt.Println("  hat <profile-name>")
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

func HandleSwitch(name string, claudeFlags []string, noRun bool) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pm := config.NewProfileManager(cfg)

	profile, err := pm.GetProfile(name)
	if err != nil {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	if err := pm.ValidateProfile(profile); err != nil {
		return fmt.Errorf("profile '%s' is corrupted: %w", name, err)
	}

	exec := executor.NewExecutor()
	if !noRun {
		if _, err := exec.FindClaude(); err != nil {
			return err
		}
	}

	cfg.CurrentProfile = name
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config with new profile: %w", err)
	}

	if err := pm.UpdateLastUsed(name); err != nil {
		ui.Warn(fmt.Sprintf("Failed to update last used timestamp: %v", err))
	}

	if res, err := pm.SyncShared(name); err != nil {
		ui.Warn(fmt.Sprintf("Failed to sync shared layer: %v", err))
	} else {
		for _, c := range res.Conflicts {
			ui.Warn(fmt.Sprintf("Shared %s conflicts with profile's own, keeping profile's", c))
		}
	}

	ui.Success(fmt.Sprintf("Switched to profile: %s", name))

	if noRun {
		ui.Info("Use 'claude' to start Claude Code with this profile.")
		return nil
	}

	if overrides := executor.AuthEnvOverrides(os.Environ()); len(overrides) > 0 {
		ui.Warn(fmt.Sprintf("%s set in environment, overriding this profile's login", strings.Join(overrides, ", ")))
	}

	ui.Info("Starting Claude Code...")
	return exec.Run(profile.Path, append(append([]string{}, profile.Metadata.CustomFlags...), claudeFlags...))
}

// Helper functions

func loadConfig() (*config.Config, error) {
	if !config.Exists() {
		return nil, ErrNotInitialized
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
