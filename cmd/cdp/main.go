package main

import (
	"fmt"
	"os"

	"github.com/tiagokriok/cdp/internal/cli"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command-line arguments
	parsed, err := cli.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Route to appropriate command handler
	switch parsed.Command {
	case "init":
		return cli.HandleInit()

	case "create":
		return cli.HandleCreate(parsed.ProfileName, parsed.Description)

	case "list":
		return cli.HandleList()

	case "delete":
		return cli.HandleDelete(parsed.ProfileName)

	case "current":
		return cli.HandleCurrent()

	case "switch":
		return cli.HandleSwitch(parsed.ProfileName, parsed.ClaudeFlags, parsed.NoRun)

	case "help":
		return cli.HandleHelp()

	case "version":
		return cli.HandleVersion(Version)

	default:
		return fmt.Errorf("unknown command: %s", parsed.Command)
	}
}
