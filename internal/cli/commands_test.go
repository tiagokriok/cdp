package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tiagokriok/cdp/internal/config"
)

// setupTestEnv creates a temporary directory structure for testing
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return tmpDir, cleanup
}

func TestHandleInit(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// First init should succeed
	err := HandleInit()
	if err != nil {
		t.Fatalf("HandleInit() failed: %v", err)
	}

	// Second init should not fail (already initialized)
	err = HandleInit()
	if err != nil {
		t.Fatalf("HandleInit() second call failed: %v", err)
	}
}

func TestHandleCreate(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize first
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Create a profile
	err = HandleCreate("work", "Work profile")
	if err != nil {
		t.Fatalf("HandleCreate() failed: %v", err)
	}

	// Try to create the same profile again (should fail)
	err = HandleCreate("work", "Duplicate")
	if err == nil {
		t.Error("HandleCreate() should fail for duplicate profile name")
	}

	// Try to create profile with invalid name
	err = HandleCreate("invalid name", "Description")
	if err == nil {
		t.Error("HandleCreate() should fail for invalid profile name")
	}
}

func TestHandleList(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// List empty profiles
	err = HandleList()
	if err != nil {
		t.Fatalf("HandleList() failed on empty: %v", err)
	}

	// Create a profile and list again
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	pm.CreateProfile("test", "Test profile")

	err = HandleList()
	if err != nil {
		t.Fatalf("HandleList() failed with profiles: %v", err)
	}
}

func TestHandleCurrent_NoProfile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Show current with no profile set
	err = HandleCurrent()
	if err != nil {
		t.Fatalf("HandleCurrent() failed: %v", err)
	}
}

func TestHandleInfo(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Create a profile
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	pm.CreateProfile("test", "Test profile")

	// Get info for existing profile
	err = HandleInfo("test")
	if err != nil {
		t.Fatalf("HandleInfo() failed: %v", err)
	}

	// Get info for non-existent profile
	err = HandleInfo("nonexistent")
	if err == nil {
		t.Error("HandleInfo() should fail for non-existent profile")
	}

	// Verify profile was created in the right place
	profileDir := filepath.Join(tmpDir, ".claude-profiles", "test")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Errorf("Profile directory not created at %s", profileDir)
	}
}

func TestHandleSwitch(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Create a profile
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	pm.CreateProfile("test", "Test profile")

	// Switch to profile with noRun=true
	err = HandleSwitch("test", []string{}, true)
	if err != nil {
		t.Fatalf("HandleSwitch() failed: %v", err)
	}

	// Verify current profile was updated
	cfg, _ = config.Load()
	if cfg.GetCurrentProfile() != "test" {
		t.Errorf("Current profile = %q, want 'test'", cfg.GetCurrentProfile())
	}

	// Switch to non-existent profile
	err = HandleSwitch("nonexistent", []string{}, true)
	if err == nil {
		t.Error("HandleSwitch() should fail for non-existent profile")
	}
}

func TestLoadConfig_NotInitialized(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Try to load config without init
	_, err := loadConfig()
	if err == nil {
		t.Error("loadConfig() should fail when not initialized")
	}
}

func TestGetConfigDirOrEmpty(t *testing.T) {
	result := getConfigDirOrEmpty()
	// Should return a non-empty string (home dir exists)
	if result == "" {
		t.Error("getConfigDirOrEmpty() should return non-empty path")
	}
}

func TestGetProfilesDirOrEmpty(t *testing.T) {
	result := getProfilesDirOrEmpty()
	// Should return a non-empty string (home dir exists)
	if result == "" {
		t.Error("getProfilesDirOrEmpty() should return non-empty path")
	}
}
