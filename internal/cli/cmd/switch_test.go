package cmd_test

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/cli"
	"github.com/tiagokriok/cdp/internal/cli/cmd"
	"github.com/tiagokriok/cdp/internal/config"
)

func TestSwitchCmd(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize config
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Create a profile to switch to
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	err = pm.CreateProfile("target-profile", "Profile for switching")
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	// Setup a dummy command for testing RunE
	switchTestCmd := &cobra.Command{
		Use:  "switch",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			claudeFlags := args[1:]
			// The noRun flag is a global persistent flag on rootCmd.
			// For direct RunE testing, we'll assume it's true or pass it explicitly.
			// Here, we hardcode true for testing purposes.
			return cli.HandleSwitch(profileName, claudeFlags, true)
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	// Test case 1: Switch to a profile without Claude flags
	err = switchTestCmd.RunE(switchTestCmd, []string{"target-profile"})
	if err != nil {
		t.Errorf("Switch to profile without flags failed: %v", err)
	}
	cfg, _ = config.Load()
	if cfg.GetCurrentProfile() != "target-profile" {
		t.Errorf("Current profile = %q, want 'target-profile'", cfg.GetCurrentProfile())
	}

	// Test case 2: Switch to a profile with Claude flags
	err = switchTestCmd.RunE(switchTestCmd, []string{"target-profile", "--continue", "--verbose"})
	if err != nil {
		t.Errorf("Switch to profile with flags failed: %v", err)
	}
	cfg, _ = config.Load()
	if cfg.GetCurrentProfile() != "target-profile" {
		t.Errorf("Current profile = %q, want 'target-profile'", cfg.GetCurrentProfile())
	}

	// Test case 3: Try to switch to a non-existent profile
	err = switchTestCmd.RunE(switchTestCmd, []string{"nonexistent-profile"})
	if err == nil {
		t.Error("Switch to non-existent profile should return an error")
	}
}

func TestSwitchCmd_UpdatesLastUsed(t *testing.T) {
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
	pm.CreateProfile("target", "Target profile")

	// Get profile before switching
	profile1, err := pm.GetProfile("target")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if !profile1.Metadata.LastUsed.IsZero() {
		t.Error("LastUsed should be zero initially")
	}

	// Switch to profile
	err = cli.HandleSwitch("target", []string{}, true)
	if err != nil {
		t.Fatalf("HandleSwitch failed: %v", err)
	}

	// Get profile after switching
	profile2, err := pm.GetProfile("target")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if profile2.Metadata.LastUsed.IsZero() {
		t.Error("LastUsed should not be zero after switch")
	}
}

func TestSwitchCmd_MultipleProfiles(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Initialize config
	err := config.Init()
	if err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)

	// Create multiple profiles
	pm.CreateProfile("profile1", "First profile")
	pm.CreateProfile("profile2", "Second profile")
	pm.CreateProfile("profile3", "Third profile")

	// Switch between them
	for _, name := range []string{"profile1", "profile2", "profile3"} {
		err := cli.HandleSwitch(name, []string{}, true)
		if err != nil {
			t.Errorf("Switch to %s failed: %v", name, err)
		}

		cfg, _ = config.Load()
		if cfg.GetCurrentProfile() != name {
			t.Errorf("Current profile = %q, want %q", cfg.GetCurrentProfile(), name)
		}
	}
}

// TestSwitchCmd_FlagPassthrough tests that Claude flags are correctly passed through
// Cobra's parsing when using the actual command structure (not bypassing with RunE directly)
func TestSwitchCmd_FlagPassthrough(t *testing.T) {
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
	pm.CreateProfile("test-profile", "Test profile")

	tests := []struct {
		name        string
		args        []string
		wantProfile string
		wantErr     bool
	}{
		{
			name:        "profile with --no-run",
			args:        []string{"switch", "test-profile", "--no-run"},
			wantProfile: "test-profile",
			wantErr:     false,
		},
		{
			name:        "profile with claude flags and --no-run",
			args:        []string{"switch", "test-profile", "--continue", "--verbose", "--no-run"},
			wantProfile: "test-profile",
			wantErr:     false,
		},
		{
			name:        "--no-run at start",
			args:        []string{"switch", "test-profile", "--no-run", "--continue"},
			wantProfile: "test-profile",
			wantErr:     false,
		},
		{
			name:    "missing profile name",
			args:    []string{"switch"},
			wantErr: true,
		},
		{
			name:    "non-existent profile",
			args:    []string{"switch", "nonexistent", "--no-run"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command state
			cmd.ResetRootCmd()

			// Get a fresh root command and execute
			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs(tt.args)

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.wantProfile != "" {
				cfg, _ := config.Load()
				if cfg.GetCurrentProfile() != tt.wantProfile {
					t.Errorf("Current profile = %q, want %q", cfg.GetCurrentProfile(), tt.wantProfile)
				}
			}
		})
	}
}
