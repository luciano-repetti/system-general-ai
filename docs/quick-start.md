# Quick Start

Install `system-general-ai`, get your AI terminal (Claude Code or Gemini CLI) configured with persona, SDD orchestrator, Engram memory, and three permission profiles. One command, interactive TUI.

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

## Interactive Installer (TUI)

Upon running `install`, a TUI will appear asking you to select the target platforms:
- **Claude Code**: Configures `~/.claude/` instances.
- **Gemini CLI**: Configures `GEMINI.md` in the current project and `~/.gemini/` for global settings.

## What `install` does in 30 seconds

1. **Platform Selection**: You choose Claude, Gemini, or both.
2. **Engram**: Downloads the Engram binary into the user-level bin dir.
3. **Claude Sync** (if selected):
   - Detects all instances.
   - Applies `permissive` profile (`bypassPermissions` + deny list).
   - Injects SDD orchestrator into `CLAUDE.md`.
   - Syncs all 11 skills.
4. **Gemini Sync** (if selected):
   - Injects SDD orchestrator into `GEMINI.md`.
   - Deploys 11 skills to `~/.gemini/skills/`.
   - Creates a `permissive.toml` policy for 100% auto-approval.

## First commands

```text
sgai install                                   # one-time, interactive
sgai sync                                      # re-apply after upgrade or new instance
sgai configure permissions <permissive|balanced|strict>
sgai configure persona <name>                  # switch the active output style (Claude)
sgai uninstall [--remove-engram]
sgai --version
```

`sgai` is interchangeable with `system-general-ai`.

## Verify the install (Claude)

| Artifact | Path |
|---|---|
| Output-style file | `<instance>/output-styles/system-general-ai.md` |
| Skill files | `<instance>/skills/<name>/SKILL.md` |
| Marker blocks | `<instance>/CLAUDE.md` contains our BEGIN/END blocks |

## Verify the install (Gemini)

| Artifact | Path |
|---|---|
| Orchestrator | `./GEMINI.md` contains our BEGIN/END blocks |
| Skill files | `~/.gemini/skills/<name>/SKILL.md` |
| Policy | `~/.gemini/policies/permissive.toml` |

## Restart your AI

Settings are read on startup. After `install`, `sync`, or `configure`, restart Claude Code or Gemini CLI to apply.
