package main

import (
	"os"
	"strings"

	"github.com/tiagokriok/hat/internal/cli/cmd"
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

func handleImplicitSwitch() {
	os.Args = rewriteImplicitSwitch(os.Args)
}

func rewriteImplicitSwitch(args []string) []string {
	if len(args) <= 1 {
		return args
	}

	knownCommands := []string{
		"init", "create", "list", "ls", "delete", "rm",
		"current", "info", "help", "version", "completion",
		"templates", "alias", "switch", "clone", "rename", "diff", "backup",
		"rule", "doctor", "mcp",
	}

	firstArg := args[1]

	if strings.HasPrefix(firstArg, "-") || strings.HasPrefix(firstArg, "__") {
		return args
	}

	for _, known := range knownCommands {
		if firstArg == known {
			return args
		}
	}

	return append([]string{args[0], "switch"}, args[1:]...)
}
