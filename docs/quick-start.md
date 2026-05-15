# Quick Start

Install `system-general-ai`, get your AI terminal (Claude Code, Gemini CLI, or Codex) configured with persona, SDD orchestrator, Engram memory, and supporting skills. One command, interactive TUI or non-interactive flags.

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
- **Codex**: Configures `~/.codex/AGENTS.md`, project `AGENTS.md`, `~/.codex/config.toml`, and `~/.codex/skills/`.

For headless or scripted installs, skip the TUI:

```bash
sgai install --target codex
sgai install --target claude,gemini,codex
sgai install --all
```

## What `install` does in 30 seconds

1. **Platform Selection**: You choose Claude, Gemini, Codex, or any combination.
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
5. **Codex Sync** (if selected):
   - Injects global rules, output style, and SDD orchestrator into `~/.codex/AGENTS.md`.
   - Injects project rules into `./AGENTS.md`.
   - Deploys skills to `~/.codex/skills/`.
   - Wires Engram in `~/.codex/config.toml`.
   - Writes Engram instruction and compact prompt files.

## First commands

```text
sgai install                                   # one-time, interactive
sgai install --target codex                    # one-time, non-interactive Codex install
sgai sync                                      # re-apply all targets after upgrade
sgai sync codex                                # re-apply only Codex
sgai configure permissions <permissive|balanced|strict>
sgai configure persona <name>                  # switch persona/output style on all targets
sgai configure persona neutral --target codex  # switch only Codex
sgai uninstall --target codex
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

## Verify the install (Codex)

| Artifact | Path |
|---|---|
| Global rules | `~/.codex/AGENTS.md` contains our BEGIN/END blocks |
| Project rules | `./AGENTS.md` contains our BEGIN/END block |
| Skill files | `~/.codex/skills/<name>/SKILL.md` |
| Engram MCP | `~/.codex/config.toml` contains `[mcp_servers.engram]` |
| Engram instructions | `~/.codex/engram-instructions.md` |
| Compact prompt | `~/.codex/engram-compact-prompt.md` |

## Restart your AI

Settings are read on startup. After `install`, `sync`, or `configure`, restart Claude Code, Gemini CLI, or Codex to apply.
