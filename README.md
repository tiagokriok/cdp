# hat: Claude Code profile switcher

A fast, lightweight CLI tool for managing multiple [Claude Code](https://claude.ai/code) profiles with seamless switching.

## Why hat?

hat solves the problem of managing different Claude Code configurations for work, personal projects, or multiple organizations. Switch between profiles instantly without manual configuration management.

## Features

- **Profile Management** - Create, list, delete, clone, rename, and compare profiles
- **Implicit Switching** - `hat work` switches and launches Claude in one step
- **Interactive TUI** - Visual profile selector with arrow key / vim navigation. Selecting a profile launches Claude (`hat --no-run` only switches)
- **Interactive Alias Wizard** - Guided TUI for setting up shell aliases with smart suggestions
- **Templates** - Pre-configured settings (restrictive / permissive) applied on creation
- **Import Existing Config** - Migrate your existing Claude Code configuration, including skills, agents, commands and plugins, into hat
- **Default Profile** - A profile symlinked to `~/.claude` runs Claude without `CLAUDE_CONFIG_DIR`, so it keeps using `~/.claude.json`
- **Shared Layer** - Share skills, agents, commands, output styles, rules and `CLAUDE.md` across profiles via symlinks
- **Plugins per Profile** - List, enable, disable, copy and diff plugins per profile
- **Directory Rules** - Bare `hat` in a mapped directory launches that profile
- **Doctor & MCP** - `hat doctor` health checks, read-only `hat mcp` listing
- **Custom Flags** - `customFlags` in a profile's `.metadata.json` are passed to Claude before your own flags
- **Backup & Restore** - Full profile backup with tar.gz compression
- **Shell Aliases** - Quick profile switching via custom shell aliases (bash/zsh/fish)
- **Shell Completion** - Auto-completion for bash, zsh, fish, and PowerShell
- **Flag Passthrough** - All Claude Code flags pass through seamlessly

## Installation

### Using `go install`

```bash
go install github.com/tiagokriok/hat/cmd/hat@latest
```

This installs `hat` to `$GOPATH/bin/` (typically `~/go/bin/`). Make sure `$GOPATH/bin` is in your PATH.

### From Source

```bash
git clone https://github.com/tiagokriok/hat.git
cd hat
make install
```

Installs to `~/.local/bin/hat`.

## Migrating from cdp

- `~/.cdp` moves to `~/.hat` automatically on the first `hat` run. Profiles are not touched.
- Old `# cdp-aliases` blocks are still listed and are replaced by `hat` ones on the next `hat alias install` (`hat alias uninstall` removes them).
- Run `make install` and delete the old binary: `rm ~/.local/bin/cdp`.
- The env var passed to Claude is now `HAT_PROFILE` (was `CLAUDE_PROFILE`).

## Quick Start

```bash
# Initialize hat
hat init

# Create profiles
hat create work "Work profile"
hat create personal "Personal projects"

# Create with a template
hat create secure --template restrictive

# Import existing Claude configuration
hat create imported --import-from ~/.config/claude-code -d "Imported config"

# Switch to a profile and run Claude
hat work --continue

# Switch without running Claude
hat work --no-run

# Interactive profile selector (no arguments), launches Claude on select
hat

# Clone, rename, compare
hat clone work work-copy
hat rename work-copy work-v2
hat diff work personal

# Backup and restore
hat backup create work
hat backup restore work-20260115-143022.tar.gz

# Set up shell aliases interactively
hat alias install
```

## Command Reference

| Command | Description |
|---|---|
| `hat init` | Initialize config and profiles directories |
| `hat create <name> [desc]` | Create a new profile |
| `hat list` / `hat ls` | List all profiles |
| `hat delete <name>` / `hat rm <name>` | Delete a profile (with confirmation) |
| `hat current` | Show currently active profile |
| `hat info [name]` | Show profile details (defaults to current) |
| `hat <profile> [flags...]` | Switch to profile and run Claude |
| `hat clone <src> <dst>` | Clone a profile |
| `hat rename <old> <new>` | Rename a profile |
| `hat diff <p1> <p2>` | Compare two profiles |
| `hat templates list` | List available templates |
| `hat templates show <name>` | Show template contents |
| `hat alias install` | Interactive alias setup wizard |
| `hat alias uninstall` | Remove all hat aliases |
| `hat alias list` | List installed aliases |
| `hat backup create <profile>` | Create tar.gz backup |
| `hat backup list` | List backups |
| `hat backup restore <file>` | Restore from backup (`--overwrite` flag) |
| `hat backup delete <file>` | Delete a backup |
| `hat shared add <kind> <path>` | Move a file/dir into `~/.hat/shared`, leave a symlink (y/N prompt) |
| `hat shared link <profile> <kinds...>` / `unlink` | Opt a profile in/out of shared kinds |
| `hat shared sync [profile]` / `list` | Re-link shared entries / show shared content |
| `hat plugins [profile]` | List plugins of a profile |
| `hat plugins enable\|disable <profile> <plugin@marketplace>` | Toggle a plugin |
| `hat plugins copy <from> <to> <plugin@marketplace>` | Copy a plugin and its marketplace to another profile |
| `hat plugins diff <a> <b>` | Compare plugins of two profiles |
| `hat rule add <path\|.> <profile>` / `list` / `rm <path>` | Manage directory rules |
| `hat doctor` | Health checks, exit 1 on failure |
| `hat mcp [profile] [--all]` / `hat mcp diff <a> <b>` | List/compare MCP servers (read-only) |
| `hat completion <shell>` | Generate shell completions (bash/zsh/fish/powershell) |
| `hat version` / `hat -v` | Show version |

### Key Flags

| Flag | Scope | Description |
|---|---|---|
| `--no-run`, `-nr` | `switch` / implicit | Switch profile without launching Claude |
| `--yes`, `-y` | all commands | Auto-confirm prompts (delete, import overwrite/proceed, `shared add`). "Remove original files?" after import stays "no". Not forwarded to Claude |
| `--template`, `-t` | `create` | Apply template (restrictive/permissive) |
| `--import-from` | `create` | Import existing Claude config directory (copies subdirectories, excludes runtime/cache/history such as `projects`, `todos`, `cache`, `history.jsonl`; symlinks skipped) |
| `--description`, `-d` | `create` | Profile description |
| `--overwrite` | `backup restore` | Overwrite existing profile on restore |

All other flags are passed through to Claude Code (e.g. `--continue`, `--verbose`, `--model`).

### Notes

- `ANTHROPIC_API_KEY`, `ANTHROPIC_AUTH_TOKEN` and `CLAUDE_CODE_OAUTH_TOKEN` in your environment override the profile's login. hat warns (names only) when any is set.
- To use your existing `~/.claude` as a profile (the default profile): `ln -s ~/.claude ~/.claude-profiles/work`. It is valid as long as the directory exists. A missing `.metadata.json` is treated as empty and created on first write (e.g. a switch). Claude itself keeps using `~/.claude.json`. Other profiles still require `.metadata.json`, `.claude.json` and `settings.json`.

### Shared layer

`hat shared add <kind> <path>` moves the path into `~/.hat/shared` and leaves a symlink (asks y/N). Kinds: `skills`, `agents`, `commands`, `output-styles`, `rules`, `CLAUDE.md`. `hat shared link <profile> <kinds...>` opts a profile in, `unlink` opts out, `list` shows content. Links are per entry, and a profile's own entry wins. `sync [profile]` prunes dangling links into shared, and also runs on every launch. Opt-in per profile.

### Plugins per profile

Plugins are scoped per profile, not symlinked. `hat plugins [profile]` lists, `enable|disable <profile> <plugin@marketplace>` toggles, `copy <from> <to> <plugin@marketplace>` copies the marketplace into `extraKnownMarketplaces` and enables it (installed on next launch), `diff <a> <b>` compares. Mods are plugins and follow the same scoping.

### Directory rules

`hat rule add <path|.> <profile>` maps a directory to a profile, `list` shows rules (`▶` marks the match for the current directory), `rm <path>` removes one. Bare `hat` in a mapped directory launches that profile (longest path wins), otherwise the TUI opens. `--no-run` is honored. Rules live under `rules:` in `~/.hat/config.yaml`.

### Doctor and MCP

`hat doctor` checks the claude binary/version, each profile (structure, `claude auth status` with a 15s timeout, broken symlinks), auth env overrides, a stale `~/.claude/.claude.json`, and rules pointing to missing profiles or paths. Exits 1 on any failure. `hat mcp [profile] [--all]` and `hat mcp diff <a> <b>` are read-only and show name, type and `scheme://host` or command, never secrets.

## Architecture

```
User → main.go (pre-parser) → Cobra CLI → Command Handler → ProfileManager / Executor → File System
```

- **Entry point** (`cmd/hat/main.go`): Pre-parses args to handle implicit switching (`hat work` becomes `hat switch work`)
- **CLI layer** (`internal/cli/`): Cobra commands, business logic handlers, Bubble Tea TUI
- **Config layer** (`internal/config/`): Global config, profile CRUD, embedded templates
- **Executor** (`internal/executor/`): Finds Claude binary, sets `CLAUDE_CONFIG_DIR` (except for the default `~/.claude` profile), execs
- **UI** (`internal/ui/`): Styled terminal output with lipgloss
- **Aliases** (`pkg/aliases/`): Shell alias management (bash/zsh/fish)
- **Backup** (`internal/backup/`): tar.gz backup/restore

### Directory Layout

```
cmd/hat/main.go          # Entry point
internal/
  cli/
    cmd/                  # Cobra subcommand definitions
    commands.go           # Command handler logic
    doctor.go, mcp.go, rules.go  # doctor, mcp, rule handlers
    interactive.go        # TUI profile selector
    interactive_alias.go  # TUI alias wizard
  config/
    config.go             # Global config (~/.hat/config.yaml)
    profile.go            # Profile operations
    copy.go               # Import/clone copier and exclusion list
    shared.go             # Shared layer (~/.hat/shared)
    plugins.go            # Per-profile plugins
    rules.go              # Directory rules
    claudejson.go         # Claude state file path
    templates.go          # Template system (embedded JSON)
    templates/            # Built-in template files
  executor/claude.go      # Claude Code execution
  backup/backup.go        # Backup/restore
  ui/output.go            # Styled output
pkg/aliases/shell.go      # Shell alias management
scripts/e2e_test.sh       # E2E test suite
```

### Configuration

**Global config** (`~/.hat/config.yaml`):

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

Each profile directory also contains `.claude.json` and `settings.json`, managed by Claude Code. hat writes `settings.json` only when applying a template.

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
bash scripts/e2e_test.sh  # Run E2E tests (48 checks)
```

## Roadmap

See [PLAN.md](PLAN.md) for the full implementation roadmap.

- **v0.1.0** - Core MVP: profile CRUD, switching, Claude execution
- **v0.2.0** - Enhanced UX: interactive TUI, styled output, Cobra migration
- **v1.0.0** - Advanced features: templates, aliases, clone, rename, diff, backup, completions
- **v1.1.0** - Interactive alias wizard with TUI
- **v1.1.1** - Bug fix: Claude flag passthrough
- **Unreleased** - Default `~/.claude` profile, subdirectory import/clone, TUI launches Claude, customFlags, auth env warning

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
