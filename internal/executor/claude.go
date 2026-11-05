package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var SystemLocations = []string{
	"/usr/local/bin/claude",
	"/usr/bin/claude",
	"/opt/homebrew/bin/claude",
}

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
	claudePath, err := e.FindClaude()
	if err != nil {
		return err
	}

	cmd := exec.Command(claudePath, flags...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLAUDE_CONFIG_DIR=%s", profilePath))

	profileName := filepath.Base(profilePath)
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLAUDE_PROFILE=%s", profileName))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute Claude Code: %w", err)
	}

	return nil
}

func (e *Executor) FindClaude() (string, error) {
	if e.claudePath != "" {
		return e.claudePath, nil
	}

	claudePath, err := exec.LookPath("claude")
	if err == nil {
		e.claudePath = claudePath
		return claudePath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("claude executable not found in PATH")
	}

	commonLocations := append(append([]string{}, SystemLocations...),
		filepath.Join(home, ".local/bin/claude"),
		filepath.Join(home, "bin/claude"),
	)

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
