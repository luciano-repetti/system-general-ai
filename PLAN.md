# system-general-ai — Implementation Plan

This document captures every stage of the project, with status tracking.

> **Persistence note**: The plan and per-stage decisions are also stored in Engram (`personal` scope, topic keys `system-general-ai/*`) so they can be recovered in any future Claude Code session, even from a different project directory.

---

## Stage 0 — Project bootstrap ✅

Initialize the repo with structure, vision and tracking.

**Done**:
- Folder structure: `docs/`, `templates/`, `scripts/`, `cmd/`, `internal/`
- README.md (vision + scope + install + commands)
- PLAN.md (this file)
- TASKS.md (snapshot for offline reference)
- Engram persistence saved

---

## Stage 1 — Architecture decisions ✅

Stack and distribution decided.

**Outputs**: `ARCHITECTURE.md` with rationale and trade-offs.

Key decisions:
- Go 1.24+
- Cobra CLI
- Bubble Tea TUI (optional, never required)
- YAML for config, JSON for `settings.json`
- Pre-compiled binaries via GitHub Releases + bootstrap install scripts
- Engram consumed upstream as a binary, not vendored or forked

---

## Stage 2 — Custom advanced system prompt ✅

Output style + global CLAUDE.md template, focused and concise.

**Outputs**:
- `templates/output-style.md` — identity + language behavior + brief response style
- `templates/CLAUDE.md.global.tmpl` — operational rules (verification, conventions, memory protocol, caching practices)

---

## Stage 3 — Claude Code templates ✅

Files installed into each instance's `.claude/`.

**Outputs**:
- `templates/CLAUDE.md.project.tmpl` — per-project skeleton with placeholders
- `templates/settings/{permissive,balanced,strict}.json` — three permission profiles
- `templates/settings/PERMISSIONS.md` — human-readable doc of each profile
- `templates/mcp/engram.json` — MCP snippet for Engram

---

## Stage 4 — SDD core ✅

Spec-Driven Development orchestrator and phase skills.

**Outputs**:
- `templates/orchestrator.md` — coordinator behavior, delegation rules, dedup gate, enforcement gate, model assignments, topic key conventions
- `templates/skills/sdd-{init,explore,propose,spec,design,tasks,apply,verify,archive}.md`

Built-in safeguards:
- Deduplication of equivalent sub-agent launches in a session
- Enforcement gate that rejects "done" claims without evidence
- Compatibility constraints for Claude Code's skill parser (flat YAML frontmatter only)

---

## Stage 5 — Engram integration ✅

Persistent memory wired up via the Engram binary.

**Outputs**:
- `internal/engram/release.go` — fetch latest release from upstream
- `internal/engram/install.go` — download, extract, place in PATH
- `internal/engram/config.go` — generate the MCP snippet

The Engram binary is consumed as-is from its upstream repo. We do not vendor or fork it.

---

## Stage 6 — Multi-Claude support ✅

First-class detection and sync across multiple Claude Code installations.

**Outputs**:
- `internal/claude/detect.go` — `CLAUDE_CONFIG_DIR` (exclusive override) + `~/.claude` (default) + `~/.claude-work*` (autodiscovery)
- `internal/claude/sync.go` — idempotent template/settings/MCP sync per instance
- `internal/claude/persona.go` — `ApplyPersona` that correctly **deletes** the `outputStyle` key on neutral (avoiding a known stale-key bug)

---

## Stage 7 — ❌ Doctor command (DEPRIORITIZED)

Skipped per direction.

---

## Stage 8 — Token-saving features ✅

Without RTK (POSIX-only, lossy compression). Solved natively with sub-agents.

**Outputs**:
- `templates/skills/shell-runner.md` — Haiku sub-agent that executes large commands, reads the full output, returns a summary. The orchestrator's context never sees the raw output.
- `templates/skills/compact-suggest.md` — surfaces `/compact` to the user when conversation accumulates
- Compacted `templates/orchestrator.md` (~25% smaller while preserving every rule)
- `internal/claude/sync.go` — skills always synced regardless of SDD on/off (so `shell-runner` and `compact-suggest` are always available)

---

## Stage 9 — Installer ✅

Single-command, cross-platform, zero-config.

**Outputs**:
- `cmd/system-general-ai/main.go` — entry point
- `internal/cli/{root,install,sync,configure,uninstall}.go` — Cobra commands
- `embed.go` — module-root embed of `templates/` into the binary
- `scripts/install.sh` — Linux/macOS bootstrap (`curl … | bash`)
- `scripts/install.ps1` — Windows bootstrap (`irm … | iex`)

`system-general-ai install` applies smart defaults without questions:
- Permissions: permissive
- SDD: ON
- shell-runner + compact-suggest: ON
- Multi-Claude detection: ON

---

## Stage 10 — Test ✅

Validation suite + token comparison.

**Outputs**:
- `internal/claude/{detect,persona,sync}_test.go` — 12 unit tests
- `internal/engram/release_test.go` — 2 unit tests
- `scripts/test-smoke.sh` — E2E in an isolated tempdir
- `scripts/test-build.sh` — cross-platform build verification
- `docs/testing.md` — guide
- `docs/token-comparison.md` — measured comparison

**Post-launch validation (real-world install on existing profile)**:
- **Bug 1 fixed** — `mergeJSONFile` did shallow top-level replacement, nuking user's `permissions` customizations on every install. Now `permissions` is deep-merged: `allow/deny/ask` are deduplicated UNION (template-first, user-only-after for stable order); `defaultMode` is user-wins. See `internal/claude/sync.go::mergePermissions` and `unionStringList`.
- **Bug 2 fixed** — skills shipped as flat `templates/skills/<name>.md` collided with the modern Claude Code directory format `skills/<name>/SKILL.md`. Restructured templates and `syncSkills` / `uninstall.go` to use directory format consistently. Uninstall preserves user-authored assets inside the dir (only removes sgai-managed `SKILL.md`).
- **Bug 3 fixed** — `scripts/test-smoke.sh` was not portable to Windows (assumed `python3` in PATH and bare `/tmp/...` paths). Added `$PY` interpreter detector and `mpath()` wrapper using `cygpath -m`. Now runs natively on Linux, macOS, and Windows.
- **New regression test** — `scripts/test-smoke.sh` now includes a second tempdir variant pre-seeded with custom `permissions.allow` entries, a non-default `permissions.defaultMode`, an unrelated user key (`theme`), and a user-authored skill in directory format. Asserts all survive install. Pins the merge contract to prevent silent regressions.
- **Permissive profile default updated** — `templates/settings/permissive.json` now sets `permissions.defaultMode: "bypassPermissions"`. Trade-off: trust the model + rely on the deny list of destructive commands. The model self-regulates by asking before aggressive actions. Balanced and strict profiles unchanged.

**Persona enhancement** (parallel work, also under Stage 2):
- `templates/output-style.md` — opening identity upgraded to "senior software architect AND systems engineer (ingeniero en sistemas), 15+ años, architecture/distributed systems/infrastructure/code-level engineering". New `## Substance` section enforces: silence beats filler, defensible claims only, trade-offs in every architectural recommendation.

---

## Stage 11 — Documentation 🔄 In Progress

Final documentation set.

**Planned outputs**:
- `docs/quick-start.md` — install + first command
- `docs/reference.md` — Engram + SDD + skills + models reference tables
- `docs/caching-and-memory.md` — best practices to maximize automatic prompt caching
- `docs/troubleshooting.md` — multi-Claude detection, persona switching, build issues

---

## Cross-stage policies

- **Conventional commits** only. No `Co-Authored-By` lines, no AI attribution.
- **Engram persistence**: every meaningful decision is saved to Engram with `personal` scope under topic keys `system-general-ai/<area>`.
- **MIT license**.

