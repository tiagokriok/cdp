package cli

import (
	"errors"
	"testing"

	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/executor"
)

func TestLoadConfig_NotInitializedSentinel(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	_, err := loadConfig()
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("loadConfig() error = %v, want ErrNotInitialized", err)
	}
}

func TestRunInteractiveMenu_InitializesWhenNotInitialized(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	if err := RunInteractiveMenu(true); err != nil {
		t.Fatalf("RunInteractiveMenu(true) failed: %v", err)
	}
	if !config.Exists() {
		t.Error("config should exist after RunInteractiveMenu on a fresh HOME")
	}
}

func TestHandleSwitch_MissingClaudeLeavesConfigUntouched(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()
	t.Setenv("PATH", "")
	t.Setenv("HOME", tmpDir)
	orig := executor.SystemLocations
	executor.SystemLocations = nil
	defer func() { executor.SystemLocations = orig }()

	if err := config.Init(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	if err := pm.CreateProfile("one", ""); err != nil {
		t.Fatal(err)
	}
	if err := pm.CreateProfile("two", ""); err != nil {
		t.Fatal(err)
	}
	if err := HandleSwitch("one", nil, true); err != nil {
		t.Fatal(err)
	}
	before, _ := pm.GetProfile("two")

	if err := HandleSwitch("two", nil, false); err == nil {
		t.Fatal("expected error when claude is missing")
	}

	cfg, _ = config.Load()
	if cfg.GetCurrentProfile() != "one" {
		t.Errorf("currentProfile = %q, want 'one'", cfg.GetCurrentProfile())
	}
	after, _ := pm.GetProfile("two")
	if after.Metadata.UsageCount != before.Metadata.UsageCount {
		t.Errorf("usageCount changed: %d -> %d", before.Metadata.UsageCount, after.Metadata.UsageCount)
	}
}
