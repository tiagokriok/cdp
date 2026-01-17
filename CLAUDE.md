# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CDP (Claude Profile Switcher) is a Go CLI tool that manages multiple Claude Code profiles, enabling seamless switching between different configurations (work, personal, multiple organizations). Version 0.2.0 provides core CRUD operations, profile switching, an interactive TUI, and Claude Code execution with the correct environment.

## Essential Commands

### Build & Install
```bash
make build              # Build binary to ./cdp
make install            # Install to ~/.local/bin/cdp
go build -ldflags="-X main.Version=0.1.0" -o cdp cmd/cdp/main.go  # Manual build with version
```

### Testing
```bash
make test               # Run all unit tests
go test -v ./...        # Run tests with verbose output
go test -v ./internal/config/          # Test specific package
go test -run TestCreateProfile ./...   # Run specific test
make coverage           # Generate HTML coverage report
bash scripts/e2e_test.sh               # Run E2E tests (24 tests)
```

### Development
```bash
make dev                # Run without building (go run)
make lint               # Format code and run go vet
go run cmd/cdp/main.go <args>  # Run with arguments
```

## Architecture

### High-Level Flow

```
User Command → `main.go` Pre-parser → Cobra CLI (Root Command/Subcommands) → Command Handler → ProfileManager/Executor → File System
                                                                                         ↓
                                                                                  Config Management
```

1. **Entry Point** (`cmd/cdp/main.go`): Performs pre-parsing to handle implicit profile switching (e.g., `cdp work` becomes `cdp switch work`). Initializes Cobra and dispatches to the root command.
2. **CLI Layer** (`internal/cli/`): Implemented using [Cobra](https://cobra.dev/) for robust command-line parsing, subcommands (defined in `internal/cli/cmd/`), and flags. Command handlers (`internal/cli/commands.go`) implement business logic. An interactive TUI (`internal/cli/interactive.go`) is launched when `cdp` is run without arguments.
3. **Config Layer** (`internal/config/`): Manages global config (`~/.cdp/config.yaml`) and profiles (`~/.claude-profiles/`)
4. **Executor** (`internal/executor/`): Executes Claude Code with correct environment variables

### Key Data Flow: Profile Switching

When running `cdp work --continue` or selecting 'work' from the TUI:
1. `main.go` pre-parser (for `cdp work`) injects `switch` command, resulting in `cdp switch work --continue`.
2. Cobra dispatches to the `switch` command (defined in `internal/cli/cmd/switch.go`).
3. `cli.HandleSwitch()` loads config, validates profile exists, and sets `currentProfile: work` in `~/.cdp/config.yaml`.
4. `cli.HandleSwitch()` updates `lastUsed` timestamp and `usageCount` in `~/.claude-profiles/work/.metadata.json`.
5. Executor sets `CLAUDE_CONFIG_DIR=/path/to/profiles/work` and runs `claude --continue`.

### Configuration System

**Two-level configuration:**
- **Global**: `~/.cdp/config.yaml` (simple: version, profilesDir, currentProfile)
- **Per-Profile**: `~/.claude-profiles/<name>/.metadata.json` (createdAt, lastUsed, description, usageCount, template, customFlags)

The ProfileManager pattern is used: handlers create `ProfileManager` with config, then call methods like `CreateProfile()`, `DeleteProfile()`, etc. This keeps config and profile operations separate from command logic.

### Profile Structure

Each profile is a directory containing:
- `.claude.json` - Claude Code OAuth config (managed by Claude, not CDP)
- `settings.json` - Claude settings (managed by Claude, not CDP)
- `.metadata.json` - CDP metadata (createdAt, lastUsed, description, usageCount, template, customFlags)

CDP only reads/writes `.metadata.json`. Other files are created as placeholders (empty `{}`).

### Parser Design

CDP now leverages the [Cobra](https://cobra.dev/) CLI framework for robust argument parsing and command dispatching.

-   **Commands**: `init`, `create`, `list` (`ls`), `delete` (`rm`), `current`, `info`, `help`, `version`, `switch` (hidden).
-   **Implicit Switch**: Any unknown first argument (not a subcommand or a flag) is treated as a profile name for an implicit `switch` command. The `main.go` pre-parser automatically injects the `switch` subcommand (e.g., `cdp work` is internally executed as `cdp switch work`).
-   **Interactive Mode**: Running `cdp` without any arguments or flags launches an interactive TUI menu for profile selection.
-   `--no-run` flag is consumed by CDP, all other flags pass through to Claude.

Example: `cdp work --continue --verbose --no-run`
-   This command is pre-parsed by `main.go` into `cdp switch work --continue --verbose --no-run`.
-   Cobra dispatches to the (hidden) `switch` command.
-   Parsed by CDP as: `{Command: "switch", ProfileName: "work", ClaudeFlags: ["--continue", "--verbose"], NoRun: true}`.

### Error Handling Pattern

Commands return errors up to `main()`, which prints to stderr and exits with code 1. No error wrapping in handlers—errors are wrapped in the layer that detects them (config, profile operations) with context using `fmt.Errorf("failed to X: %w", err)`.

### Testing Strategy

- **Unit tests**: Table-driven tests in `*_test.go` files, focus on config and parser
- **E2E tests**: `scripts/e2e_test.sh` tests full lifecycle in isolated temp directory
- **No mocking**: Tests use `t.TempDir()` and real file operations with temporary HOME override

## Important Constraints

1. **Profile name validation**: Only alphanumeric, hyphens, underscores, max 50 chars (see `profileNamePattern` regex in `internal/config/profile.go`)
2. **Cannot delete current profile**: Must switch to another profile first (enforced in `DeleteProfile()`)
3. **Claude executable detection**: Searches PATH first, then common locations (`/usr/local/bin/claude`, `~/.local/bin/claude`, etc.)
4. **Version injection**: Version is set via ldflags at build time, defaults to "dev"

## Future Architecture Notes (Tier 2 & 3)

The current architecture is designed to accommodate future enhancements:
- **Tier 2**: Cobra (CLI framework) and Bubble Tea (TUI) are implemented. The custom parser has been replaced and core command handlers (`internal/cli/commands.go`) remain.
- **Tier 3**: Templates, shell aliases, backup/restore will extend `ProfileManager` with new methods. Config struct will be expanded but backwards compatible with current YAML structure.

See PLAN.md for detailed roadmap.

# Git Instructions

- **Commit Messages**: Use imperative mood and concise descriptions, only commit files you modified or deleted or created
- **Branch Naming**: Use `feature/` or `bugfix/` or `docs/` or `refactor/` or `chore/` or `test/` prefixes
- **Pull Requests**: Create PRs from feature branches to `main`
- **Git Hooks**: Use pre-commit hooks for linting and formatting
- **Conventional Commits**: Follow the Conventional Commits specification https://www.conventionalcommits.org/en/v1.0.0/
