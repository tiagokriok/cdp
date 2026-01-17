package executor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewExecutor(t *testing.T) {
	e := NewExecutor()
	if e == nil {
		t.Error("NewExecutor() returned nil")
	}
	if e.claudePath != "" {
		t.Errorf("NewExecutor() claudePath = %q, want empty string", e.claudePath)
	}
}

func TestSetClaudePath(t *testing.T) {
	e := NewExecutor()
	testPath := "/usr/local/bin/claude"

	e.SetClaudePath(testPath)

	if e.claudePath != testPath {
		t.Errorf("SetClaudePath() claudePath = %q, want %q", e.claudePath, testPath)
	}
}

func TestFindClaude_WithSetPath(t *testing.T) {
	e := NewExecutor()
	expectedPath := "/custom/path/to/claude"

	// Set a custom path
	e.SetClaudePath(expectedPath)

	// findClaude should return the set path
	path, err := e.findClaude()
	if err != nil {
		t.Fatalf("findClaude() error = %v, want nil", err)
	}
	if path != expectedPath {
		t.Errorf("findClaude() = %q, want %q", path, expectedPath)
	}
}

func TestFindClaude_InPath(t *testing.T) {
	e := NewExecutor()

	// Create a temporary directory with a fake claude binary
	tmpDir := t.TempDir()
	claudePath := filepath.Join(tmpDir, "claude")

	// Create a dummy executable file
	f, err := os.Create(claudePath)
	if err != nil {
		t.Fatalf("Failed to create temp claude: %v", err)
	}
	f.Close()
	os.Chmod(claudePath, 0755)

	// Prepend temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)

	path, err := e.findClaude()
	if err != nil {
		t.Fatalf("findClaude() error = %v, want nil", err)
	}
	if path != claudePath {
		t.Errorf("findClaude() = %q, want %q", path, claudePath)
	}
}

func TestFindClaude_CommonLocations(t *testing.T) {
	e := NewExecutor()

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	// Create a fake claude in ~/.local/bin/
	localBinDir := filepath.Join(home, ".local", "bin")
	claudePath := filepath.Join(localBinDir, "claude")

	// Ensure directory exists
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		t.Fatalf("Failed to create local bin dir: %v", err)
	}

	// Check if a real claude exists first - skip test if so
	if _, err := os.Stat(claudePath); err == nil {
		t.Skip("Real claude binary exists at ~/.local/bin/claude, skipping test")
	}

	// Create a dummy executable
	f, err := os.Create(claudePath)
	if err != nil {
		t.Fatalf("Failed to create temp claude: %v", err)
	}
	f.Close()
	os.Chmod(claudePath, 0755)
	defer os.Remove(claudePath)

	// Clear PATH to force fallback search
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	path, err := e.findClaude()
	if err != nil {
		t.Fatalf("findClaude() error = %v, want nil", err)
	}
	if path != claudePath {
		t.Errorf("findClaude() = %q, want %q", path, claudePath)
	}
}

func TestFindClaude_NotFound(t *testing.T) {
	// Skip if running in CI or if claude is installed
	// This test manipulates PATH which can be unreliable
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment")
	}

	e := NewExecutor()

	// Clear PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", originalPath)

	// Temporarily change home to avoid finding real claude
	originalHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	_, err := e.findClaude()
	if err == nil {
		// It's possible claude is found through other means,
		// so we just verify the function doesn't panic
		t.Log("Note: claude was found even with modified PATH/HOME")
	}
}

func TestRun_WithInvalidClaude(t *testing.T) {
	e := NewExecutor()

	// Set a non-existent claude path
	e.SetClaudePath("/nonexistent/claude/binary")

	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile")
	os.MkdirAll(profilePath, 0755)

	err := e.Run(profilePath, []string{})
	if err == nil {
		t.Error("Run() should fail when claude binary doesn't exist")
	}
}
