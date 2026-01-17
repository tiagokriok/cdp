package cmd_test

import (
	"os"
	"testing"
)

// setupTestEnv creates a temporary directory structure for testing
// and returns a cleanup function.
func setupTestEnv(t *testing.T) func() {
	t.Helper()

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return cleanup
}

// mockStdin temporarily replaces os.Stdin with a buffer containing the provided input.
// It returns a function that restores the original os.Stdin.
func mockStdin(input string) func() {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	_ = w.Close()
	os.Stdin = r
	return func() {
		os.Stdin = oldStdin
	}
}