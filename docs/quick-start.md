# Quick Start

Install `system-general-ai`, get a Claude Code instance configured with persona, SDD orchestrator, Engram memory, and three permission profiles. One command, no questions.

## Install

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/lucianorepetti/system-general-ai/main/scripts/install.sh | bash
```

### Windows (PowerShell 5.1+)

```powershell
irm https://raw.githubusercontent.com/lucianorepetti/system-general-ai/main/scripts/install.ps1 | iex
```

The bootstrap detects OS + arch, places the binary under your user-level PATH (`~/.local/bin` on POSIX, `%LOCALAPPDATA%\system-general-ai\bin` on Windows), creates the `sgai` short alias, and runs `system-general-ai install` automatically.

## What `install` does in 30 seconds

`internal/cli/install.go` does this with no prompts:

1. Detect every Claude Code instance: `CLAUDE_CONFIG_DIR` (exclusive override) or `~/.claude` + `~/.claude-work*`
2. Download the Engram binary into the user-level bin dir
3. For each detected instance, apply these defaults:

| Setting | Default |
|---|---|
| Permissions profile | `permissive` (`bypassPermissions` + destructive-command deny list) |
| SDD orchestrator block in `CLAUDE.md` | ON |
| `shell-runner` skill | always installed |
| `compact-suggest` skill | always installed |
| Multi-Claude detection | ON |
| Engram MCP entry in `settings.json` | wired to the binary path |
| Output style | `system-general-ai` |

Re-running is idempotent. Marker-delimited sections in `CLAUDE.md` are merged in place (`mergeMarkedSection` in `internal/claude/sync.go:206`). The `permissions` block is deep-merged so user `allow`/`deny`/`ask` entries survive (`mergePermissions`, `internal/claude/sync.go:309`).

## First commands

```text
sgai install                                   # one-time, zero-config
sgai sync                                      # re-apply after upgrade or new instance
sgai configure permissions <permissive|balanced|strict>
sgai configure persona <name>                  # switch the active output style
sgai configure persona -                       # remove outputStyle (neutral)
sgai configure persona neutral                 # same as "-"
sgai uninstall [--remove-engram]
sgai --version
```

`sgai` is interchangeable with `system-general-ai`.

`configure persona -` and `configure persona neutral` both DELETE the `outputStyle` key from `settings.json` — they do not set it to an empty string. See `internal/claude/persona.go:32`.

## Verify the install

For each instance under `~/.claude` (or wherever it was synced), check:

| Artifact | Path |
|---|---|
| Output-style file | `<instance>/output-styles/system-general-ai.md` |
| Skill files | `<instance>/skills/<name>/SKILL.md` (11 names: `sdd-*`, `shell-runner`, `compact-suggest`) |
| Marker block | `<instance>/CLAUDE.md` contains `<!-- BEGIN: system-general-ai/global-rules -->` |
| Marker block (SDD) | `<instance>/CLAUDE.md` contains `<!-- BEGIN: system-general-ai/sdd-orchestrator -->` |
| Engram MCP entry | `<instance>/settings.json` has `mcpServers.engram.command` pointing at the engram binary |
| Persona | `<instance>/settings.json` has `"outputStyle": "system-general-ai"` |

## Restart Claude Code

Settings are read on Claude Code startup. After `install`, `sync`, or `configure`, restart Claude Code to apply.

