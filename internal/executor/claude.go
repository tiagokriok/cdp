package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Executor handles execution of Claude Code
type Executor struct {
	claudePath string
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	return &Executor{}
}

// Run executes Claude Code with the specified profile
func (e *Executor) Run(profilePath string, flags []string) error {
	// Find Claude executable
	claudePath, err := e.findClaude()
	if err != nil {
		return err
	}

	// Build command
	cmd := exec.Command(claudePath, flags...)

	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLAUDE_CONFIG_DIR=%s", profilePath))

	// Extract profile name from path
	profileName := filepath.Base(profilePath)
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLAUDE_PROFILE=%s", profileName))

	// Inherit stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute
	if err := cmd.Run(); err != nil {
		// Check if it's an exit error (Claude returned non-zero)
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Return the same exit code
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute Claude Code: %w", err)
	}

	return nil
}

// findClaude locates the Claude executable
func (e *Executor) findClaude() (string, error) {
	if e.claudePath != "" {
		return e.claudePath, nil
	}

	// Try to find claude in PATH first
	claudePath, err := exec.LookPath("claude")
	if err == nil {
		e.claudePath = claudePath
		return claudePath, nil
	}

	// Fallback: search in common locations
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("claude executable not found in PATH")
	}

	commonLocations := []string{
		"/usr/local/bin/claude",
		"/usr/bin/claude",
		"/opt/homebrew/bin/claude",
		filepath.Join(home, ".local/bin/claude"),
		filepath.Join(home, "bin/claude"),
	}

	for _, location := range commonLocations {
		if _, err := os.Stat(location); err == nil {
			e.claudePath = location
			return location, nil
		}
	}

	return "", fmt.Errorf("claude executable not found. Please ensure Claude Code is installed and in your PATH")
}

// SetClaudePath manually sets the path to the Claude executable
func (e *Executor) SetClaudePath(path string) {
	e.claudePath = path
}
