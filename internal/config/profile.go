package config

import (
	"encoding/json"
	"fmt"
	"github.com/tiagokriok/hat/internal/prompt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
	UsageCount  int       `json:"usageCount"`
	Template    string    `json:"template,omitempty"`
	CustomFlags []string  `json:"customFlags,omitempty"`
	Shared      []string  `json:"shared,omitempty"`
}

// Profile represents a Claude Code profile
type Profile struct {
	Name     string
	Path     string
	Metadata ProfileMetadata
}

func IsDefaultClaudeDir(path string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	def, err := filepath.EvalSymlinks(filepath.Join(home, ".claude"))
	if err != nil {
		return false
	}
	return resolved == def
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
	return pm.CreateProfileWithTemplate(name, description, "")
}

// CreateProfileWithTemplate creates a new profile with an optional template
func (pm *ProfileManager) CreateProfileWithTemplate(name, description, template string) error {
	if err := ValidateName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	profilePath := filepath.Join(pm.config.ProfilesDir, name)

	// Check if profile already exists
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// Validate template if specified
	if template != "" {
		tm := NewTemplateManager()
		if !tm.TemplateExists(template) {
			return fmt.Errorf("template '%s' not found", template)
		}
	}

	// Create profile directory
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Create metadata
	metadata := ProfileMetadata{
		CreatedAt:   time.Now(),
		Description: description,
		UsageCount:  0,
		Template:    template,
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

	// Apply template if specified
	if template != "" {
		tm := NewTemplateManager()
		if err := tm.ApplyTemplate(profilePath, template); err != nil {
			os.RemoveAll(profilePath)
			return fmt.Errorf("failed to apply template: %w", err)
		}
	}

	return nil
}

// DeleteProfile deletes a profile
func (pm *ProfileManager) DeleteProfile(name string) error {
	if err := ValidateName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	profilePath := filepath.Join(pm.config.ProfilesDir, name)

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
		name := entry.Name()
		profilePath := filepath.Join(pm.config.ProfilesDir, name)
		if info, err := os.Stat(profilePath); err != nil || !info.IsDir() {
			continue
		}

		metadata, err := pm.loadMetadata(profilePath)
		if err != nil {
			continue
		}

		profiles = append(profiles, Profile{
			Name:     name,
			Path:     profilePath,
			Metadata: metadata,
		})
	}

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

	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile '%s' does not exist", name)
	}

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
	profile.Metadata.UsageCount++
	return pm.saveMetadata(profile.Path, profile.Metadata)
}

// ValidateProfile checks if a profile directory structure is valid
func (pm *ProfileManager) ValidateProfile(profile *Profile) error {
	if _, err := os.Stat(profile.Path); os.IsNotExist(err) {
		return fmt.Errorf("profile directory does not exist")
	}

	if IsDefaultClaudeDir(profile.Path) {
		return nil
	}

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
		if os.IsNotExist(err) && IsDefaultClaudeDir(profilePath) {
			return ProfileMetadata{}, nil
		}
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
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}

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

// CloneProfile clones an existing profile to a new name
func (pm *ProfileManager) CloneProfile(source, dest string) error {
	if err := ValidateName(dest); err != nil {
		return fmt.Errorf("invalid destination name: %w", err)
	}

	sourceProfile, err := pm.GetProfile(source)
	if err != nil {
		return fmt.Errorf("source profile '%s' does not exist", source)
	}

	destPath := filepath.Join(pm.config.ProfilesDir, dest)

	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("profile '%s' already exists", dest)
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if _, _, err := copyProfileTree(sourceProfile.Path, destPath); err != nil {
		os.RemoveAll(destPath)
		return fmt.Errorf("failed to copy profile: %w", err)
	}

	if err := rewritePluginPaths(sourcePrefixes(sourceProfile.Path), destPath); err != nil {
		os.RemoveAll(destPath)
		return err
	}

	newMetadata := sourceProfile.Metadata
	newMetadata.CreatedAt = time.Now()
	newMetadata.LastUsed = time.Time{}
	newMetadata.UsageCount = 0
	newMetadata.Description = fmt.Sprintf("Cloned from %s", source)

	if err := pm.saveMetadata(destPath, newMetadata); err != nil {
		os.RemoveAll(destPath)
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// RenameProfile renames an existing profile
func (pm *ProfileManager) RenameProfile(oldName, newName string) error {
	if err := ValidateName(newName); err != nil {
		return fmt.Errorf("invalid new name: %w", err)
	}

	// Check old exists
	if !pm.ProfileExists(oldName) {
		return fmt.Errorf("profile '%s' does not exist", oldName)
	}

	// Check new doesn't exist
	if pm.ProfileExists(newName) {
		return fmt.Errorf("profile '%s' already exists", newName)
	}

	// Can't rename current profile while it's active
	if pm.config.CurrentProfile == oldName {
		return fmt.Errorf("cannot rename the current profile '%s', switch to another profile first", oldName)
	}

	oldPath := filepath.Join(pm.config.ProfilesDir, oldName)
	newPath := filepath.Join(pm.config.ProfilesDir, newName)

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename profile: %w", err)
	}

	return nil
}

// ImportProfile imports an existing Claude configuration into a new profile
func (pm *ProfileManager) ImportProfile(sourcePath, name, description string) error {
	if err := ValidateName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	if strings.HasPrefix(sourcePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		sourcePath = filepath.Join(home, sourcePath[2:])
	} else if sourcePath == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		sourcePath = home
	}

	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	sourceInfo, err := os.Stat(absSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source path '%s' does not exist", sourcePath)
		}
		return fmt.Errorf("failed to access source path: %w", err)
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("source path '%s' is not a directory", sourcePath)
	}

	destPath := filepath.Join(pm.config.ProfilesDir, name)
	profileExists := false
	if _, err := os.Stat(destPath); err == nil {
		profileExists = true
	}

	entries, err := os.ReadDir(absSourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	var (
		foundClaudeConfig bool
		foundSettings     bool
		foundMetadata     bool
		otherFiles        []string
		copiedDirs        []string
		excluded          []string
		topSymlinks       []string
	)

	for _, entry := range entries {
		if entry.Name() == MetadataFileName {
			foundMetadata = true
			continue
		}
		if IsExcludedEntry(entry.Name()) {
			excluded = append(excluded, entry.Name())
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			topSymlinks = append(topSymlinks, entry.Name())
			continue
		}
		if entry.IsDir() {
			copiedDirs = append(copiedDirs, entry.Name())
			continue
		}

		switch entry.Name() {
		case ClaudeConfigFile:
			foundClaudeConfig = true
		case ClaudeSettingsFile:
			foundSettings = true
		default:
			otherFiles = append(otherFiles, entry.Name())
		}
	}

	fmt.Println()
	fmt.Println("ℹ Import Preview")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Source:      %s\n", absSourcePath)
	fmt.Printf("Destination: %s\n", destPath)
	fmt.Printf("Profile:     %s\n", name)
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}
	fmt.Println(strings.Repeat("─", 60))

	fmt.Println()
	fmt.Println("Files to import:")
	fileCount := 0
	if foundClaudeConfig {
		fmt.Printf("  ✓ %s (OAuth credentials)\n", ClaudeConfigFile)
		fileCount++
	} else {
		fmt.Printf("  ✗ %s (will create empty placeholder)\n", ClaudeConfigFile)
	}
	if foundSettings {
		fmt.Printf("  ✓ %s (Claude settings)\n", ClaudeSettingsFile)
		fileCount++
	} else {
		fmt.Printf("  ✗ %s (will create empty placeholder)\n", ClaudeSettingsFile)
	}
	if foundMetadata {
		fmt.Printf("  ~ %s (will overwrite with new hat metadata)\n", MetadataFileName)
	}
	for _, fileName := range otherFiles {
		fmt.Printf("  ✓ %s\n", fileName)
		fileCount++
	}

	if len(copiedDirs) > 0 {
		fmt.Printf("\nSubdirectories to copy: %s\n", strings.Join(copiedDirs, ", "))
	}
	if len(excluded) > 0 {
		fmt.Printf("Excluded (runtime/cache/history): %s\n", strings.Join(excluded, ", "))
	}
	if len(topSymlinks) > 0 {
		fmt.Printf("Symlinks skipped: %s\n", strings.Join(topSymlinks, ", "))
	}

	fmt.Printf("\nTop-level files to copy: %d\n", fileCount)

	if !foundClaudeConfig {
		fmt.Println("\n⚠ Warning: No .claude.json found - this may not be a valid Claude configuration directory")
	}

	if profileExists {
		fmt.Printf("\n⚠ Profile '%s' already exists!\n", name)
		if !prompt.Ask("Overwrite existing profile?", false) {
			return fmt.Errorf("import cancelled by user")
		}
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("failed to remove existing profile: %w", err)
		}
	}

	if !prompt.Ask("\nProceed with import?", true) {
		return fmt.Errorf("import cancelled by user")
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	fmt.Println("\nImporting files...")
	copiedCount, symlinks, err := copyProfileTree(absSourcePath, destPath)
	if err != nil {
		os.RemoveAll(destPath)
		return fmt.Errorf("failed to copy profile: %w", err)
	}
	if err := rewritePluginPaths(sourcePrefixes(absSourcePath), destPath); err != nil {
		os.RemoveAll(destPath)
		return err
	}
	fmt.Printf("  ✓ Copied %d files\n", copiedCount)
	for _, link := range symlinks {
		fmt.Printf("  ~ Skipped symlink %s\n", link)
	}

	if !foundClaudeConfig {
		claudeConfigPath := filepath.Join(destPath, ClaudeConfigFile)
		if err := os.WriteFile(claudeConfigPath, []byte("{}"), 0644); err != nil {
			os.RemoveAll(destPath)
			return fmt.Errorf("failed to create Claude config file: %w", err)
		}
		fmt.Printf("  + Created empty %s\n", ClaudeConfigFile)
	}

	if !foundSettings {
		settingsPath := filepath.Join(destPath, ClaudeSettingsFile)
		if err := os.WriteFile(settingsPath, []byte("{}"), 0644); err != nil {
			os.RemoveAll(destPath)
			return fmt.Errorf("failed to create settings file: %w", err)
		}
		fmt.Printf("  + Created empty %s\n", ClaudeSettingsFile)
	}

	metadata := ProfileMetadata{
		CreatedAt:   time.Now(),
		Description: description,
		UsageCount:  0,
		Template:    "",
	}

	if foundMetadata {
		importedMetadata, err := pm.loadMetadata(absSourcePath)
		if err == nil {
			metadata.Template = importedMetadata.Template
			metadata.CustomFlags = importedMetadata.CustomFlags
		}
	}

	if err := pm.saveMetadata(destPath, metadata); err != nil {
		os.RemoveAll(destPath)
		return fmt.Errorf("failed to save metadata: %w", err)
	}
	fmt.Printf("  ✓ Created %s\n", MetadataFileName)

	fmt.Printf("\nℹ Original files are still in: %s\n", absSourcePath)
	if prompt.Ask("Remove original files?", false) {
		if err := os.RemoveAll(absSourcePath); err != nil {
			fmt.Printf("⚠ Failed to remove original files: %v\n", err)
			fmt.Println("You can manually delete them later")
		} else {
			fmt.Printf("  ✓ Removed original files from %s\n", absSourcePath)
		}
	}

	return nil
}
