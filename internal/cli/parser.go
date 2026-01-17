package cli

import (
	"fmt"
	"strings"
)

// ParsedArgs represents parsed command-line arguments
type ParsedArgs struct {
	Command     string   // Command to execute: init, create, list, delete, current, switch, help, version
	ProfileName string   // Profile name for create, delete, switch commands
	Description string   // Description for create command
	ClaudeFlags []string // Flags to pass through to Claude
	NoRun       bool     // For switch command: don't run Claude after switching
}

// Parse parses command-line arguments
func Parse(args []string) (*ParsedArgs, error) {
	if len(args) == 0 {
		return &ParsedArgs{Command: "help"}, nil
	}

	parsed := &ParsedArgs{}

	switch args[0] {
	case "init":
		parsed.Command = "init"
		return parsed, nil

	case "list", "ls":
		parsed.Command = "list"
		return parsed, nil

	case "current":
		parsed.Command = "current"
		return parsed, nil

	case "help", "--help", "-h":
		parsed.Command = "help"
		return parsed, nil

	case "version", "--version", "-v":
		parsed.Command = "version"
		return parsed, nil

	case "create":
		if len(args) < 2 {
			return nil, fmt.Errorf("create command requires a profile name")
		}
		parsed.Command = "create"
		parsed.ProfileName = args[1]

		// Optional description (rest of arguments joined)
		if len(args) > 2 {
			parsed.Description = strings.Join(args[2:], " ")
		}
		return parsed, nil

	case "delete", "del", "rm":
		if len(args) < 2 {
			return nil, fmt.Errorf("delete command requires a profile name")
		}
		parsed.Command = "delete"
		parsed.ProfileName = args[1]
		return parsed, nil

	default:
		// Treat first argument as profile name (switch command)
		parsed.Command = "switch"
		parsed.ProfileName = args[0]
		parsed.ClaudeFlags = []string{} // Initialize to empty slice

		// Parse remaining arguments for Claude flags and --no-run
		for i := 1; i < len(args); i++ {
			if args[i] == "--no-run" {
				parsed.NoRun = true
			} else {
				parsed.ClaudeFlags = append(parsed.ClaudeFlags, args[i])
			}
		}

		return parsed, nil
	}
}

// IsValidCommand checks if a command is valid
func IsValidCommand(cmd string) bool {
	validCommands := map[string]bool{
		"init":    true,
		"create":  true,
		"list":    true,
		"delete":  true,
		"current": true,
		"switch":  true,
		"help":    true,
		"version": true,
	}
	return validCommands[cmd]
}

// GetCommandHelp returns help text for a specific command
func GetCommandHelp(cmd string) string {
	help := map[string]string{
		"init":    "Initialize CDP configuration",
		"create":  "Create a new profile: cdp create <name> [description]",
		"list":    "List all profiles",
		"delete":  "Delete a profile: cdp delete <name>",
		"current": "Show current active profile",
		"switch":  "Switch to profile and run Claude: cdp <profile> [flags]",
		"help":    "Show this help message",
		"version": "Show version information",
	}
	return help[cmd]
}

// GetUsage returns the usage string
func GetUsage() string {
	return `CDP - Claude Profile Switcher

USAGE:
    cdp <command> [arguments]

COMMANDS:
    init                     Initialize CDP configuration
    create <name> [desc]     Create a new profile
    list                     List all profiles
    delete <name>            Delete a profile
    current                  Show current active profile
    <profile> [flags]        Switch to profile and run Claude
    help                     Show this help message
    version                  Show version information

EXAMPLES:
    # Initialize CDP
    cdp init

    # Create profiles
    cdp create work "Work profile for SmartTalks"
    cdp create personal

    # List profiles
    cdp list

    # Switch and run Claude
    cdp work --continue
    cdp personal --verbose

    # Switch without running Claude
    cdp work --no-run

    # Show current profile
    cdp current

    # Delete a profile
    cdp delete old-profile

For more information, visit: https://github.com/tiagokriok/cdp`
}
