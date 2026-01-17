package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

const (
	MetadataFileName   = ".metadata.json"
	ClaudeConfigFile   = ".claude.json"
	ClaudeSettingsFile = "settings.json"
)

var (
	// Valid profile name pattern: alphanumeric, hyphens, underscores
	profileNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ProfileMetadata contains metadata about a profile
type ProfileMetadata struct {
	CreatedAt   time.Time `json:"createdAt"`
	LastUsed    time.Time `json:"lastUsed,omitempty"`
	Description string    `json:"description,omitempty"`
}

// Profile represents a Claude Code profile
type Profile struct {
	Name     string
	Path     string
	Metadata ProfileMetadata
}

// ProfileManager handles profile operations
type ProfileManager struct {
	config *Config
}

// NewProfileManager creates a new profile manager
func NewProfileManager(cfg *Config) *ProfileManager {
	return &ProfileManager{config: cfg}
}

// ValidateName checks if a profile name is valid
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if len(name) > 50 {
		return fmt.Errorf("profile name too long (max 50 characters)")
	}
	if !profileNamePattern.MatchString(name) {
		return fmt.Errorf("invalid profile name: use only letters, numbers, hyphens, and underscores")
	}
	return nil
}

// CreateProfile creates a new profile
func (pm *ProfileManager) CreateProfile(name, description string) error {
	if err := ValidateName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	profilePath := filepath.Join(pm.config.ProfilesDir, name)

	// Check if profile already exists
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// Create profile directory
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Create metadata
	metadata := ProfileMetadata{
		CreatedAt:   time.Now(),
		Description: description,
	}

	if err := pm.saveMetadata(profilePath, metadata); err != nil {
		// Clean up on failure
		os.RemoveAll(profilePath)
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	// Create empty Claude config files as placeholders
	claudeConfigPath := filepath.Join(profilePath, ClaudeConfigFile)
	if err := os.WriteFile(claudeConfigPath, []byte("{}"), 0644); err != nil {
		os.RemoveAll(profilePath)
		return fmt.Errorf("failed to create Claude config file: %w", err)
	}

	settingsPath := filepath.Join(profilePath, ClaudeSettingsFile)
	if err := os.WriteFile(settingsPath, []byte("{}"), 0644); err != nil {
		os.RemoveAll(profilePath)
		return fmt.Errorf("failed to create settings file: %w", err)
	}

	return nil
}

// DeleteProfile deletes a profile
func (pm *ProfileManager) DeleteProfile(name string) error {
	if err := ValidateName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	profilePath := filepath.Join(pm.config.ProfilesDir, name)

	// Check if profile exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Don't allow deleting the current profile
	if pm.config.CurrentProfile == name {
		return fmt.Errorf("cannot delete the current profile '%s', switch to another profile first", name)
	}

	// Remove the profile directory
	if err := os.RemoveAll(profilePath); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// ListProfiles lists all profiles
func (pm *ProfileManager) ListProfiles() ([]Profile, error) {
	entries, err := os.ReadDir(pm.config.ProfilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Profile{}, nil
		}
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []Profile
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		profilePath := filepath.Join(pm.config.ProfilesDir, name)

		// Load metadata
		metadata, err := pm.loadMetadata(profilePath)
		if err != nil {
			// Skip profiles with invalid metadata
			continue
		}

		profiles = append(profiles, Profile{
			Name:     name,
			Path:     profilePath,
			Metadata: metadata,
		})
	}

	// Sort by last used (most recent first), then by name
	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].Metadata.LastUsed.IsZero() && profiles[j].Metadata.LastUsed.IsZero() {
			return profiles[i].Name < profiles[j].Name
		}
		if profiles[i].Metadata.LastUsed.IsZero() {
			return false
		}
		if profiles[j].Metadata.LastUsed.IsZero() {
			return true
		}
		return profiles[i].Metadata.LastUsed.After(profiles[j].Metadata.LastUsed)
	})

	return profiles, nil
}

// GetProfile retrieves a profile by name
func (pm *ProfileManager) GetProfile(name string) (*Profile, error) {
	if err := ValidateName(name); err != nil {
		return nil, fmt.Errorf("invalid profile name: %w", err)
	}

	profilePath := filepath.Join(pm.config.ProfilesDir, name)

	// Check if profile exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile '%s' does not exist", name)
	}

	// Load metadata
	metadata, err := pm.loadMetadata(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile metadata: %w", err)
	}

	return &Profile{
		Name:     name,
		Path:     profilePath,
		Metadata: metadata,
	}, nil
}

// UpdateLastUsed updates the last used timestamp for a profile
func (pm *ProfileManager) UpdateLastUsed(name string) error {
	profile, err := pm.GetProfile(name)
	if err != nil {
		return err
	}

	profile.Metadata.LastUsed = time.Now()
	return pm.saveMetadata(profile.Path, profile.Metadata)
}

// ValidateProfile checks if a profile directory structure is valid
func (pm *ProfileManager) ValidateProfile(profile *Profile) error {
	// Check if directory exists
	if _, err := os.Stat(profile.Path); os.IsNotExist(err) {
		return fmt.Errorf("profile directory does not exist")
	}

	// Check if required files exist
	requiredFiles := []string{MetadataFileName, ClaudeConfigFile, ClaudeSettingsFile}
	for _, file := range requiredFiles {
		filePath := filepath.Join(profile.Path, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file '%s' is missing", file)
		}
	}

	return nil
}

// loadMetadata loads profile metadata from disk
func (pm *ProfileManager) loadMetadata(profilePath string) (ProfileMetadata, error) {
	metadataPath := filepath.Join(profilePath, MetadataFileName)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return ProfileMetadata{}, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata ProfileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return ProfileMetadata{}, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}

// saveMetadata saves profile metadata to disk
func (pm *ProfileManager) saveMetadata(profilePath string, metadata ProfileMetadata) error {
	metadataPath := filepath.Join(profilePath, MetadataFileName)

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// ProfileExists checks if a profile exists
func (pm *ProfileManager) ProfileExists(name string) bool {
	profilePath := filepath.Join(pm.config.ProfilesDir, name)
	_, err := os.Stat(profilePath)
	return err == nil
}
