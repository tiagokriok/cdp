# CDP - Claude Profile Switcher

A fast, lightweight CLI tool for managing multiple Claude Code profiles with seamless switching.

## Why CDP?

CDP solves the problem of managing different Claude Code configurations for work, personal projects, or multiple organizations. Switch between profiles instantly without manual configuration management.

## Features (v1.0.0)

- **Profile Management**: Create, list, delete, clone, and rename profiles
- **Interactive TUI**: Visual profile selector with arrow key navigation
- **Templates**: Pre-configured settings templates (restrictive/permissive)
- **Shell Aliases**: Quick profile switching via shell aliases
- **Backup/Restore**: Full profile backup with tar.gz compression
- **Profile Diff**: Compare settings between two profiles
- **Shell Completion**: Auto-completion for bash, zsh, fish, and PowerShell
- **Flag Passthrough**: All Claude Code flags pass through seamlessly

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

# Create a profile with a template
cdp create secure-work "Secure work profile" --template restrictive

# List profiles
cdp list

# Switch to a profile and run Claude
cdp work --continue

# Switch without running Claude
cdp work --no-run

# Show current profile
cdp current

# Interactive profile selector (no arguments)
cdp

# Clone a profile
cdp clone work work-backup

# Compare two profiles
cdp diff work personal

# Backup a profile
cdp backup create work

# Delete a profile
cdp delete personal
```

## Command Reference

### `cdp init`
Initialize CDP configuration directory (`~/.cdp/`) and profiles directory (`~/.claude-profiles/`).

### `cdp create <name> [description]`
Create a new profile with the given name and optional description.

**Flags:**
- `--template, -t <name>`: Apply a settings template (restrictive, permissive)

Examples:
```bash
cdp create work "Work profile for SmartTalks"
cdp create secure-work "Secure profile" --template restrictive
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

### `cdp clone <source> <destination>`
Clone an existing profile to create a new one with the same settings.

Example:
```bash
cdp clone work work-backup
```

### `cdp rename <old-name> <new-name>`
Rename an existing profile. Cannot rename the currently active profile.

Example:
```bash
cdp rename old-work work
```

### `cdp diff <profile1> <profile2>`
Compare two profiles and show differences in their settings.

Example:
```bash
cdp diff work personal
```

### `cdp templates`
Manage settings templates.

**Subcommands:**
- `cdp templates list`: List available templates
- `cdp templates show <name>`: Show template contents

Examples:
```bash
cdp templates list
cdp templates show restrictive
```

**Built-in Templates:**
- **restrictive**: Disabled auto-updates, requires confirmation for file changes
- **permissive**: Full permissions with auto-updates enabled

### `cdp alias`
Manage shell aliases for quick profile switching.

**Subcommands:**
- `cdp alias install`: Install shell aliases for all profiles
- `cdp alias uninstall`: Remove all CDP aliases from shell config
- `cdp alias list`: List currently installed aliases

Examples:
```bash
# Install aliases (adds to .bashrc/.zshrc/config.fish)
cdp alias install

# After installation, use aliases directly:
# cdp-work is equivalent to "cdp work"
cdp-work --continue

# List installed aliases
cdp alias list

# Remove all aliases
cdp alias uninstall
```

### `cdp backup`
Backup and restore profiles.

**Subcommands:**
- `cdp backup create <profile>`: Create a tar.gz backup of a profile
- `cdp backup list`: List all available backups
- `cdp backup restore <file>`: Restore a profile from backup
- `cdp backup delete <file>`: Delete a backup file

**Flags for restore:**
- `--overwrite`: Overwrite existing profile if it exists

Examples:
```bash
# Create a backup
cdp backup create work
# Output: Backup created: ~/.cdp/backups/work-20240115-143022.tar.gz

# List backups
cdp backup list

# Restore a backup
cdp backup restore work-20240115-143022.tar.gz

# Restore and overwrite existing
cdp backup restore work-20240115-143022.tar.gz --overwrite

# Delete a backup
cdp backup delete work-20240115-143022.tar.gz
```

### `cdp completion`
Generate shell completion scripts.

**Subcommands:**
- `cdp completion bash`: Generate bash completion script
- `cdp completion zsh`: Generate zsh completion script
- `cdp completion fish`: Generate fish completion script
- `cdp completion powershell`: Generate PowerShell completion script

Examples:
```bash
# Bash (add to ~/.bashrc)
source <(cdp completion bash)

# Zsh (add to ~/.zshrc)
source <(cdp completion zsh)

# Fish
cdp completion fish | source

# PowerShell
cdp completion powershell | Out-String | Invoke-Expression
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

### Tier 1: Core MVP (v0.1.0) ✅
- ✅ Profile CRUD operations
- ✅ Profile switching
- ✅ Claude Code execution with environment

### Tier 2: Enhanced UX (v0.2.0) ✅
- ✅ Interactive TUI menu
- ✅ Styled output with colors
- ✅ Profile info command

### Tier 3: Advanced Features (v1.0.0) ✅
- ✅ Profile templates (restrictive/permissive)
- ✅ Shell aliases (bash/zsh/fish)
- ✅ Clone and rename profiles
- ✅ Profile diff comparison
- ✅ Backup/restore functionality
- ✅ Shell auto-completion

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
