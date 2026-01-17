# CDP - Claude Profile Switcher

A fast, lightweight CLI tool for managing multiple Claude Code profiles with seamless switching.

## Why CDP?

CDP solves the problem of managing different Claude Code configurations for work, personal projects, or multiple organizations. Switch between profiles instantly without manual configuration management.

## Features (v0.2.0)

- Interactive TUI menu for profile selection
- Create, list, and delete profiles
- Switch between profiles instantly
- Execute Claude Code with the correct environment
- Pass through all Claude Code flags
- Simple, intuitive command-line interface

## Installation

### From Source

```bash
git clone https://github.com/tiagokriok/cdp.git
cd cdp
make install
```

This installs `cdp` to `~/.local/bin/cdp`. Make sure `~/.local/bin` is in your PATH.

## Quick Start

```bash
# Initialize CDP
cdp init

# Create profiles
cdp create work "Work profile for SmartTalks"
cdp create personal "Personal projects"

# List profiles
cdp list

# Switch to a profile and run Claude
cdp work --continue

# Switch without running Claude
cdp work --no-run

# Show current profile
cdp current

# Delete a profile
cdp delete personal
```

## Command Reference

### `cdp init`
Initialize CDP configuration directory (`~/.cdp/`) and profiles directory (`~/.claude-profiles/`).

### `cdp create <name> [description]`
Create a new profile with the given name and optional description.

Example:
```bash
cdp create work "Work profile for SmartTalks"
```

### `cdp list`
List all available profiles with their metadata.

### `cdp delete <name>`
Delete a profile. You'll be prompted for confirmation.

Example:
```bash
cdp delete work
```

### `cdp current`
Show the currently active profile.

### `cdp info [profile-name]`
Show detailed information about a profile. If no profile name is provided, it shows information for the current active profile.

Example:
```bash
cdp info work
cdp info
```

### `cdp <profile> [flags...]`
Switch to the specified profile and run Claude Code with any provided flags.

Examples:
```bash
# Switch and run Claude in continue mode
cdp work --continue

# Switch and run with verbose output
cdp work --verbose

# Switch without running Claude
cdp work --no-run
```

## Profile Directory Structure

Profiles are stored in `~/.claude-profiles/`:

```
~/.claude-profiles/
├── work/
│   ├── .claude.json       # Claude Code OAuth config
│   ├── settings.json      # Claude settings
│   └── .metadata.json     # CDP metadata (createdAt, lastUsed, description, usageCount, template, customFlags)
└── personal/
    ├── .claude.json
    ├── settings.json
    └── .metadata.json
```

## Configuration

CDP configuration is stored in `~/.cdp/config.yaml`:

```yaml
version: "1.0"
profilesDir: /home/user/.claude-profiles
currentProfile: work
```

## Development

### Prerequisites

- Go 1.21 or later
- Claude Code CLI installed

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Coverage Report

```bash
make coverage
```

## Roadmap

See [PLAN.md](PLAN.md) for the full implementation roadmap.

### Tier 2: Enhanced UX (v0.2.0)
- ✅ Interactive TUI menu
- ✅ Styled output with colors
- ✅ Profile info command

### Tier 3: Advanced Features (v1.0.0)
- Profile templates
- Shell aliases
- Clone and diff profiles
- Backup/restore functionality

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
