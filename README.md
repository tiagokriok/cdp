# CDP - Claude Profile Switcher

A fast, lightweight CLI tool for managing multiple [Claude Code](https://claude.ai/code) profiles with seamless switching.

## Why CDP?

CDP solves the problem of managing different Claude Code configurations for work, personal projects, or multiple organizations. Switch between profiles instantly without manual configuration management.

## Features

- **Profile Management** - Create, list, delete, clone, rename, and compare profiles
- **Implicit Switching** - `cdp work` switches and launches Claude in one step
- **Interactive TUI** - Visual profile selector with arrow key / vim navigation
- **Interactive Alias Wizard** - Guided TUI for setting up shell aliases with smart suggestions
- **Templates** - Pre-configured settings (restrictive / permissive) applied on creation
- **Import Existing Config** - Migrate your existing Claude Code configuration into CDP
- **Backup & Restore** - Full profile backup with tar.gz compression
- **Shell Aliases** - Quick profile switching via custom shell aliases (bash/zsh/fish)
- **Shell Completion** - Auto-completion for bash, zsh, fish, and PowerShell
- **Flag Passthrough** - All Claude Code flags pass through seamlessly

## Installation

### Using `go install`

```bash
go install github.com/tiagokriok/cdp/cmd/cdp@latest
```

This installs `cdp` to `$GOPATH/bin/` (typically `~/go/bin/`). Make sure `$GOPATH/bin` is in your PATH.

### From Source

```bash
git clone https://github.com/tiagokriok/cdp.git
cd cdp
make install
```

Installs to `~/.local/bin/cdp`.

## Quick Start

```bash
# Initialize CDP
cdp init

# Create profiles
cdp create work "Work profile"
cdp create personal "Personal projects"

# Create with a template
cdp create secure --template restrictive

# Import existing Claude configuration
cdp create imported --import-from ~/.config/claude-code -d "Imported config"

# Switch to a profile and run Claude
cdp work --continue

# Switch without running Claude
cdp work --no-run

# Interactive profile selector (no arguments)
cdp

# Clone, rename, compare
cdp clone work work-copy
cdp rename work-copy work-v2
cdp diff work personal

# Backup and restore
cdp backup create work
cdp backup restore work-20260115-143022.tar.gz

# Set up shell aliases interactively
cdp alias install
```

## Command Reference

| Command | Description |
|---|---|
| `cdp init` | Initialize config and profiles directories |
| `cdp create <name> [desc]` | Create a new profile |
| `cdp list` / `cdp ls` | List all profiles |
| `cdp delete <name>` / `cdp rm <name>` | Delete a profile (with confirmation) |
| `cdp current` | Show currently active profile |
| `cdp info [name]` | Show profile details (defaults to current) |
| `cdp <profile> [flags...]` | Switch to profile and run Claude |
| `cdp clone <src> <dst>` | Clone a profile |
| `cdp rename <old> <new>` | Rename a profile |
| `cdp diff <p1> <p2>` | Compare two profiles |
| `cdp templates list` | List available templates |
| `cdp templates show <name>` | Show template contents |
| `cdp alias install` | Interactive alias setup wizard |
| `cdp alias uninstall` | Remove all CDP aliases |
| `cdp alias list` | List installed aliases |
| `cdp backup create <profile>` | Create tar.gz backup |
| `cdp backup list` | List backups |
| `cdp backup restore <file>` | Restore from backup (`--overwrite` flag) |
| `cdp backup delete <file>` | Delete a backup |
| `cdp completion <shell>` | Generate shell completions (bash/zsh/fish/powershell) |
| `cdp version` / `cdp -v` | Show version |

### Key Flags

| Flag | Scope | Description |
|---|---|---|
| `--no-run`, `-nr` | `switch` / implicit | Switch profile without launching Claude |
| `--template`, `-t` | `create` | Apply template (restrictive/permissive) |
| `--import-from` | `create` | Import existing Claude config directory |
| `--description`, `-d` | `create` | Profile description |
| `--overwrite` | `backup restore` | Overwrite existing profile on restore |

All other flags are passed through to Claude Code (e.g. `--continue`, `--verbose`, `--model`).

## Architecture

```
User → main.go (pre-parser) → Cobra CLI → Command Handler → ProfileManager / Executor → File System
```

- **Entry point** (`cmd/cdp/main.go`): Pre-parses args to handle implicit switching (`cdp work` becomes `cdp switch work`)
- **CLI layer** (`internal/cli/`): Cobra commands, business logic handlers, Bubble Tea TUI
- **Config layer** (`internal/config/`): Global config, profile CRUD, embedded templates
- **Executor** (`internal/executor/`): Finds Claude binary, sets `CLAUDE_CONFIG_DIR`, execs
- **UI** (`internal/ui/`): Styled terminal output with lipgloss
- **Aliases** (`pkg/aliases/`): Shell alias management (bash/zsh/fish)
- **Backup** (`internal/backup/`): tar.gz backup/restore

### Directory Layout

```
cmd/cdp/main.go          # Entry point
internal/
  cli/
    cmd/                  # Cobra subcommand definitions
    commands.go           # Command handler logic
    interactive.go        # TUI profile selector
    interactive_alias.go  # TUI alias wizard
  config/
    config.go             # Global config (~/.cdp/config.yaml)
    profile.go            # Profile operations
    templates.go          # Template system (embedded JSON)
    templates/            # Built-in template files
  executor/claude.go      # Claude Code execution
  backup/backup.go        # Backup/restore
  ui/output.go            # Styled output
pkg/aliases/shell.go      # Shell alias management
scripts/e2e_test.sh       # E2E test suite
```

### Configuration

**Global config** (`~/.cdp/config.yaml`):

```yaml
version: "1.0"
profilesDir: /home/user/.claude-profiles
currentProfile: work
```

**Per-profile metadata** (`~/.claude-profiles/<name>/.metadata.json`):

```json
{
  "createdAt": "2026-01-16T22:15:00Z",
  "lastUsed": "2026-01-17T14:30:00Z",
  "description": "Work profile",
  "usageCount": 42,
  "template": "restrictive",
  "customFlags": ["--verbose"]
}
```

Each profile directory also contains `.claude.json` (OAuth) and `settings.json` (Claude settings), both managed by Claude Code itself.

## Dependencies

| Package | Purpose |
|---|---|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework and command routing |
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | Interactive TUI framework |
| [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) | TUI components (text input) |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling and colors |
| [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) | YAML config parsing |

## Development

### Prerequisites

- Go 1.25+
- Claude Code CLI installed

### Build & Test

```bash
make build              # Build binary
make install            # Install to ~/.local/bin
make test               # Run unit tests
make coverage           # HTML coverage report
make lint               # Format + vet
make dev                # Run without building
bash scripts/e2e_test.sh  # Run E2E tests
```

## Roadmap

See [PLAN.md](PLAN.md) for the full implementation roadmap.

- **v0.1.0** - Core MVP: profile CRUD, switching, Claude execution
- **v0.2.0** - Enhanced UX: interactive TUI, styled output, Cobra migration
- **v1.0.0** - Advanced features: templates, aliases, clone, rename, diff, backup, completions
- **v1.1.0** - Interactive alias wizard with TUI
- **v1.1.1** - Bug fix: Claude flag passthrough

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
