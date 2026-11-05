package cmd_test

import (
	"github.com/tiagokriok/hat/internal/prompt"
	"os"
	"strings"
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

func mockStdin(input string) func() {
	return prompt.SetInput(strings.NewReader(input))
}
