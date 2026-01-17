package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/executor"
)

// HandleInit initializes the CDP configuration
func HandleInit() error {
	if config.Exists() {
		fmt.Println("CDP is already initialized.")
		fmt.Println("Configuration directory:", getConfigDirOrEmpty())
		fmt.Println("Profiles directory:", getProfilesDirOrEmpty())
		return nil
	}

	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to initialize CDP: %w", err)
	}

	fmt.Println("CDP initialized successfully!")
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

	fmt.Printf("Profile '%s' created successfully!\n", name)
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

	if len(profiles) == 0 {
		fmt.Println("No profiles found.")
		fmt.Println("\nCreate a profile:")
		fmt.Println("  cdp create <name> [description]")
		return nil
	}

	fmt.Printf("Found %d profile(s):\n\n", len(profiles))

	currentProfile := cfg.GetCurrentProfile()

	for _, profile := range profiles {
		// Mark current profile
		marker := " "
		if profile.Name == currentProfile {
			marker = "*"
		}

		fmt.Printf("%s %s\n", marker, profile.Name)

		if profile.Metadata.Description != "" {
			fmt.Printf("  Description: %s\n", profile.Metadata.Description)
		}

		fmt.Printf("  Created: %s\n", formatTime(profile.Metadata.CreatedAt))

		if !profile.Metadata.LastUsed.IsZero() {
			fmt.Printf("  Last used: %s\n", formatTime(profile.Metadata.LastUsed))
		}

		fmt.Println()
	}

	if currentProfile != "" {
		fmt.Printf("Current profile: %s\n", currentProfile)
	}

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
		fmt.Println("Deletion cancelled.")
		return nil
	}

	if err := pm.DeleteProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	fmt.Printf("Profile '%s' deleted successfully.\n", name)
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
		fmt.Println("No profile is currently active.")
		fmt.Println("\nSwitch to a profile:")
		fmt.Println("  cdp <profile-name>")
		return nil
	}

	fmt.Printf("Current profile: %s\n", currentProfile)

	// Get profile details
	pm := config.NewProfileManager(cfg)
	profile, err := pm.GetProfile(currentProfile)
	if err != nil {
		// Profile might have been deleted
		fmt.Println("(Warning: Profile directory not found)")
		return nil
	}

	if profile.Metadata.Description != "" {
		fmt.Printf("Description: %s\n", profile.Metadata.Description)
	}

	fmt.Printf("Location: %s\n", profile.Path)

	if !profile.Metadata.LastUsed.IsZero() {
		fmt.Printf("Last used: %s\n", formatTime(profile.Metadata.LastUsed))
	}

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

	// Update current profile
	if err := cfg.SetCurrentProfile(name); err != nil {
		return fmt.Errorf("failed to set current profile: %w", err)
	}

	// Update last used timestamp
	if err := pm.UpdateLastUsed(name); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to update last used timestamp: %v\n", err)
	}

	fmt.Printf("Switched to profile: %s\n", name)

	if noRun {
		fmt.Println("Use 'claude' to start Claude Code with this profile.")
		return nil
	}

	// Run Claude Code
	fmt.Println("Starting Claude Code...")
	exec := executor.NewExecutor()
	return exec.Run(profile.Path, claudeFlags)
}

// HandleHelp shows help information
func HandleHelp() error {
	fmt.Println(GetUsage())
	return nil
}

// HandleVersion shows version information
func HandleVersion(version string) error {
	if version == "" {
		version = "dev"
	}
	fmt.Printf("CDP (Claude Profile Switcher) version %s\n", version)
	return nil
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

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02")
	}
}
