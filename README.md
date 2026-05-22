# system-general-ai

A focused, minimal AI ecosystem configurator **for Claude Code, Gemini CLI, and Codex**.

## What it does

`system-general-ai` configures your AI terminal with:

- A custom **system prompt** (output style) — short, direct, no pedagogical voice, neutral Spanish + English
- **Spec-Driven Development** workflow — orchestrator + 9 phase skills with per-phase model assignment
- **Engram** persistent memory across sessions
- A **`shell-runner`** sub-agent that absorbs large command outputs (test runs, builds, log dumps) so they never inflate the main context
- A **`compact-suggest`** skill that nudges the user to run `/compact` when the conversation accumulates
- **Multi-Platform support** — Install for Claude Code, Gemini CLI, Codex, or any combination via an interactive TUI or non-interactive flags.
- **Three permission profiles** (permissive default, balanced, strict) — one flag to switch
- A **single-command zero-config installer** for Windows, macOS, and Linux

## Why

If you use Claude Code, Gemini CLI, or Codex daily and want the productivity layer without:

- A chatty pedagogical persona that adds tokens to every turn
- An installer that requires cloning a repo, aliasing binaries, and migrating configs across multiple credentials
- Adapters for agents you don't use (Cursor, OpenCode, etc.)
- Components you'll never touch

…then this is the wrapper for you. It does one thing — configure your AI well — and gets out of the way.

## Install

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/luciano-repetti/system-general-ai/main/scripts/install.sh | bash
```

### Windows (PowerShell 5.1+)

```powershell
irm https://raw.githubusercontent.com/luciano-repetti/system-general-ai/main/scripts/install.ps1 | iex
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
system-general-ai install --target codex
system-general-ai install --all
system-general-ai sync [claude|gemini|codex...]
system-general-ai configure permissions <permissive|balanced|strict>
system-general-ai configure persona <name|"-"|"neutral"> [--target claude|gemini|codex]
system-general-ai uninstall [--target claude|gemini|codex] [--remove-engram]
system-general-ai --version
```

`sgai` is interchangeable with `system-general-ai`.

Use `install` for first-time setup. Use `sync` after upgrading `system-general-ai`, after editing templates from source, or after manual config changes you want to repair.

## Defaults applied at install

| Setting | Default | Why |
|---|---|---|
| Permissions profile | `permissive` | Productivity-first, with destructive-command denylist |
| SDD orchestrator | ON | Available by default; the agent uses the smallest useful SDD path |
| `shell-runner` | ON | Cross-platform; saves tokens on every large command |
| `compact-suggest` | ON | Surfaces `/compact` proactively when history grows |
| Engram | Installed and wired | Persistent memory across sessions |
| Output style | `system-general-ai` | Custom direct-and-technical persona |
| Gemini Policy | `permissive` | YOLO mode with auto-approval for shell commands |
| Codex SDD | ON | SDD via `AGENTS.md` + skills, relaxed for direct answers and trivial work |
| Codex Engram | ON | MCP + instruction/compact prompt files |

Reconfigure any of them post-install with `system-general-ai configure ...` — never required.

## Project structure

```text
cmd/system-general-ai/       Binary entry point
internal/cli/                Cobra commands (install, sync, configure, uninstall)
internal/claude/             Claude Code detection + sync
internal/gemini/             Gemini CLI sync logic
internal/codex/              Codex detection + sync
internal/engram/             Engram binary download + MCP config generation
templates/                   Embedded into the binary at build time
scripts/                     Bootstrap install scripts (bash + PowerShell)
docs/                        Documentation
```

## Build from source

Requires **Go 1.24+**.

```bash
git clone https://github.com/luciano-repetti/system-general-ai
cd system-general-ai
go mod tidy
go build -o sgai ./cmd/system-general-ai
./sgai install
```

## Development workflow

Templates are embedded at build time from `templates/`. If you change prompts, skills, settings, or Codex/Gemini adapters:

```bash
go test ./internal/...
go build -o sgai ./cmd/system-general-ai
./sgai sync codex      # or: ./sgai sync claude gemini codex
```

Then restart the target AI tool. `sync` is idempotent and preserves user content outside managed marker blocks.

## Release workflow

The one-line installers download the latest GitHub Release. To publish a new version:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions builds these assets automatically:

```text
system-general-ai_<tag>_windows_amd64.zip
system-general-ai_<tag>_windows_arm64.zip
system-general-ai_<tag>_linux_amd64.tar.gz
system-general-ai_<tag>_linux_arm64.tar.gz
system-general-ai_<tag>_darwin_amd64.tar.gz
system-general-ai_<tag>_darwin_arm64.tar.gz
checksums.txt
```

The Windows `irm ... | iex` installer expects the Windows `.zip` names above and `system-general-ai.exe` at the archive root.

For cross-platform builds:

```bash
./scripts/test-build.sh
```

## Testing

```bash
go test ./internal/...     # unit tests
./scripts/test-smoke.sh    # E2E in isolated tempdirs, never touches real ~/.claude or ~/.codex
./scripts/test-build.sh    # cross-platform compile check
```

See `docs/testing.md` for details.

## License

MIT.

