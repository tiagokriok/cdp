package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tiagokriok/cdp/internal/config"
)

var SystemLocations = []string{
	"/usr/local/bin/claude",
	"/usr/bin/claude",
	"/opt/homebrew/bin/claude",
}

type Executor struct {
	claudePath string
}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Run(profilePath string, flags []string) error {
	claudePath, err := e.FindClaude()
	if err != nil {
		return err
	}

	cmd := exec.Command(claudePath, flags...)

	cmd.Env = BuildEnv(os.Environ(), profilePath, config.IsDefaultClaudeDir(profilePath))

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

func BuildEnv(base []string, profilePath string, isDefault bool) []string {
	env := make([]string, 0, len(base)+2)
	for _, kv := range base {
		if strings.HasPrefix(kv, "CLAUDE_CONFIG_DIR=") || strings.HasPrefix(kv, "CLAUDE_PROFILE=") {
			continue
		}
		env = append(env, kv)
	}
	if !isDefault {
		env = append(env, "CLAUDE_CONFIG_DIR="+profilePath)
	}
	return append(env, "CLAUDE_PROFILE="+filepath.Base(profilePath))
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

func (e *Executor) SetClaudePath(path string) {
	e.claudePath = path
}
