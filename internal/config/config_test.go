package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test initialization
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify config directory was created
	configDir := filepath.Join(tmpDir, ConfigDirName)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created")
	}

	// Verify profiles directory was created
	profilesDir := filepath.Join(tmpDir, ProfilesDirName)
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		t.Errorf("Profiles directory was not created")
	}

	// Verify config file was created
	configPath := filepath.Join(configDir, ConfigFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize first
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify default values
	if cfg.Version != ConfigVersion {
		t.Errorf("Expected version %s, got %s", ConfigVersion, cfg.Version)
	}

	expectedProfilesDir := filepath.Join(tmpDir, ProfilesDirName)
	if cfg.ProfilesDir != expectedProfilesDir {
		t.Errorf("Expected profiles dir %s, got %s", expectedProfilesDir, cfg.ProfilesDir)
	}

	// Modify and save
	cfg.CurrentProfile = "test-profile"
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load again and verify changes persisted
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Second Load() failed: %v", err)
	}

	if cfg2.CurrentProfile != "test-profile" {
		t.Errorf("Expected current profile 'test-profile', got '%s'", cfg2.CurrentProfile)
	}
}

func TestSetCurrentProfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Set current profile
	if err := cfg.SetCurrentProfile("my-profile"); err != nil {
		t.Fatalf("SetCurrentProfile() failed: %v", err)
	}

	// Verify it was saved
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg2.GetCurrentProfile() != "my-profile" {
		t.Errorf("Expected current profile 'my-profile', got '%s'", cfg2.GetCurrentProfile())
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Should not exist initially
	if Exists() {
		t.Error("Exists() returned true before initialization")
	}

	// Initialize
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Should exist now
	if !Exists() {
		t.Error("Exists() returned false after initialization")
	}
}

func TestGetConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ConfigDirName)
	if configDir != expected {
		t.Errorf("Expected config dir %s, got %s", expected, configDir)
	}
}

func TestGetDefaultProfilesDir(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	profilesDir, err := GetDefaultProfilesDir()
	if err != nil {
		t.Fatalf("GetDefaultProfilesDir() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ProfilesDirName)
	if profilesDir != expected {
		t.Errorf("Expected profiles dir %s, got %s", expected, profilesDir)
	}
}
