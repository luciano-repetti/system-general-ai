# system-general-ai

A focused, minimal AI ecosystem configurator **for Claude Code and Gemini CLI**.

## What it does

`system-general-ai` configures your AI terminal with:

- A custom **system prompt** (output style) — short, direct, no pedagogical voice, neutral Spanish + English
- **Spec-Driven Development** workflow — orchestrator + 9 phase skills with per-phase model assignment
- **Engram** persistent memory across sessions
- A **`shell-runner`** sub-agent that absorbs large command outputs (test runs, builds, log dumps) so they never inflate the main context
- A **`compact-suggest`** skill that nudges the user to run `/compact` when the conversation accumulates
- **Multi-Platform support** — Install for Claude Code, Gemini CLI, or both via an interactive TUI.
- **Three permission profiles** (permissive default, balanced, strict) — one flag to switch
- A **single-command zero-config installer** for Windows, macOS, and Linux

## Why

If you use Claude Code or Gemini CLI daily and want the productivity layer without:

- A chatty pedagogical persona that adds tokens to every turn
- An installer that requires cloning a repo, aliasing binaries, and migrating configs across multiple credentials
- Adapters for agents you don't use (Cursor, OpenCode, etc.)
- Components you'll never touch

…then this is the wrapper for you. It does one thing — configure your AI well — and gets out of the way.

## Install

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/lucianorepetti/system-general-ai/main/scripts/install.sh | bash
```

### Windows (PowerShell 5.1+)

```powershell
irm https://raw.githubusercontent.com/lucianorepetti/system-general-ai/main/scripts/install.ps1 | iex
```

Both scripts:

1. Detect OS + architecture
2. Download the matching binary from GitHub Releases
3. Place it under your user-level PATH (`~/.local/bin/` on POSIX, `%LOCALAPPDATA%\system-general-ai\bin\` on Windows)
4. Create a `sgai` short alias
5. Run `system-general-ai install` automatically with an interactive TUI to choose your platform.

Total time: under a minute, no admin rights needed.

## Commands

```text
system-general-ai install
system-general-ai sync
system-general-ai configure permissions <permissive|balanced|strict>
system-general-ai configure persona <name|"-"|"neutral">
system-general-ai uninstall [--remove-engram]
system-general-ai --version
```

`sgai` is interchangeable with `system-general-ai`.

## Defaults applied at install

| Setting | Default | Why |
|---|---|---|
| Permissions profile | `permissive` | Productivity-first, with destructive-command denylist |
| SDD orchestrator | ON | Always available; the agent decides when to delegate |
| `shell-runner` | ON | Cross-platform; saves tokens on every large command |
| `compact-suggest` | ON | Surfaces `/compact` proactively when history grows |
| Engram | Installed and wired | Persistent memory across sessions |
| Output style | `system-general-ai` | Custom direct-and-technical persona |
| Gemini Policy | `permissive` | YOLO mode with auto-approval for shell commands |

Reconfigure any of them post-install with `system-general-ai configure ...` — never required.

## Project structure

```text
cmd/system-general-ai/       Binary entry point
internal/cli/                Cobra commands (install, sync, configure, uninstall)
internal/claude/             Claude Code detection + sync
internal/gemini/             Gemini CLI sync logic
internal/engram/             Engram binary download + MCP config generation
templates/                   Embedded into the binary at build time
scripts/                     Bootstrap install scripts (bash + PowerShell)
docs/                        Documentation
```

## Build from source

Requires **Go 1.24+**.

```bash
git clone https://github.com/lucianorepetti/system-general-ai
cd system-general-ai
go mod tidy
go build -o sgai ./cmd/system-general-ai
./sgai install
```

For cross-platform builds:

```bash
./scripts/test-build.sh
```

## Testing

```bash
go test ./internal/...     # unit tests
./scripts/test-smoke.sh    # E2E in an isolated tempdir, never touches your real ~/.claude
./scripts/test-build.sh    # cross-platform compile check
```

See `docs/testing.md` for details.

## License

MIT.

