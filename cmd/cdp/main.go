package main

import (
	"os"
	"strings"

	"github.com/tiagokriok/cdp/internal/cli/cmd"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
)

func main() {
	cmd.Version = Version
	handleImplicitSwitch()
	cmd.Execute()
}

// handleImplicitSwitch checks if the user is trying to run a switch command
// without explicitly typing 'switch'. e.g., `cdp work`.
// If so, it injects the 'switch' command into the arguments for Cobra.
func handleImplicitSwitch() {
	if len(os.Args) <= 1 {
		return
	}

	// List of all user-facing commands and their aliases.
	// This determines if the first argument is a command or a profile name.
	knownCommands := []string{
		"init", "create", "list", "ls", "delete", "rm",
		"current", "info", "help", "version", "completion",
		"templates", "alias", "switch", "clone", "rename", "diff", "backup",
	}

	firstArg := os.Args[1]

	// If the first argument is a flag, it's not a profile name.
	if strings.HasPrefix(firstArg, "-") {
		return
	}

	// If the first argument is a known command, it's not a profile name.
	for _, known := range knownCommands {
		if firstArg == known {
			return
		}
	}

	// If we're here, the first argument is not a flag and not a known command.
	// We assume it's a profile name for the 'switch' command.
	// We inject 'switch' into the os.Args slice.
	originalArgs := os.Args
	os.Args = append([]string{originalArgs[0], "switch"}, originalArgs[1:]...)
}
