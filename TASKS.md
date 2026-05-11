# system-general-ai — Task Snapshot

> Live tracking lives in Claude Code's task system + Engram. This file is a static snapshot for offline reference.
>
> Last refresh: 2026-05-04 (after real-world install validation against an existing profile).

## Stage status

| Stage | Status | Notes |
|---|---|---|
| 0. Project bootstrap | ✅ Done | Folder, README, PLAN, TASKS created |
| 1. Architecture decisions | ✅ Done | Go 1.24+, Cobra, embedded templates, single binary, Engram consumed upstream |
| 2. Custom system prompt | ✅ Done | `output-style.md` (architect + systems engineer persona, neutral Spanish, substance-only) + global CLAUDE.md template |
| 3. Claude Code templates | ✅ Done | Per-project CLAUDE.md, three permission profiles, MCP snippets |
| 4. SDD core | ✅ Done | Orchestrator + 9 phase skills, dedup gate, enforcement gate, model assignments |
| 5. Engram integration | ✅ Done | Binary download + MCP config generator |
| 6. Multi-Claude support | ✅ Done | `CLAUDE_CONFIG_DIR` exclusive override + `~/.claude` default + `~/.claude-work*` autodiscovery; persona switch correctly deletes `outputStyle` key |
| 7. Doctor command | ❌ Skipped | Deprioritized per direction |
| 8. Token-saving features | ✅ Done | `shell-runner` and `compact-suggest` skills (no RTK — sub-agent pattern instead); orchestrator compacted ~25% |
| 9. Installer | ✅ Done | Single-command zero-config installers for bash + PowerShell, embedded templates, smart defaults |
| 10. Test | ✅ Done + post-launch fixes | See "Post-launch findings" below |
| 11. Documentation | 🔄 In Progress | quick-start, reference, caching-and-memory, troubleshooting |

## Post-launch findings (closed during real-world validation against `~/.claude-work2`)

| # | Bug | Status | Fix |
|---|---|---|---|
| 1 | `mergeJSONFile` shallow-replaced `permissions`, nuking user customizations | ✅ Fixed | `mergePermissions` deep-merge: `allow/deny/ask` UNION; `defaultMode` user-wins |
| 2 | Skills shipped as flat `<name>.md`, collided with modern `<name>/SKILL.md` directory format | ✅ Fixed | Restructured templates + sync logic to use directory format consistently |
| 3 | `scripts/test-smoke.sh` was not portable to Windows | ✅ Fixed | Added `$PY` interpreter detector and `mpath()` `cygpath -m` wrapper |
| 4 | Smoke test only ran against empty tempdir — couldn't catch merge regressions | ✅ Fixed | New variant pre-seeds populated `settings.json` + user skill, asserts they survive |
| 5 | `permissive` profile defaulted to ask-mode | ✅ Changed | `defaultMode: "bypassPermissions"` (trust model + rely on deny list of destructive commands) |

Environmental quirk noted (not a sgai bug):
- Multi-profile setups using NTFS hardlinks across `CLAUDE.md` files cause one install to propagate to all linked profiles. sgai is unaware. Break the hardlink before installing if isolation matters.

## Upstream issues addressed

| # | Title | Stage |
|---|---|---|
| #204 | Neutral persona still uses Rioplatense output style | 6 |
| #253 | CLAUDE_CONFIG_DIR for multi-profile installations | 6 |
| #194 | Orchestrator runs multiple agents with same task | 4 |
| #262 | SDD does not enforce TDD/verify/task completion | 4 |
| #365 | SDD slash commands fail on Claude Code v2.1.113+ | 4 |
| #347 | RTK component for 60-90% shell output reduction | 8 (rejected — sub-agent pattern instead) |
| #111 | RTK token compression as ecosystem component | 8 (rejected — sub-agent pattern instead) |
| #259 | Deduplicate engram-protocol, lazy-load orchestrator | 8 |
| #177 | Engram fails to install on Windows | 9 |
| #246 | Installer auto-detect broken | 9 |
| #201 | Update fails on Windows 10 | 9 |
| #366 | Windows install bugs | 9 |
| #378 | Missing dependencies during install | 9 |
| #95 | TUI crash without TTY | 9 |
| #114 | Installer ignores state.json | 9 |

## Recovery instructions

If you start a fresh Claude Code session and need to resume work on this project:

1. Open Claude Code in this folder: `C:\Users\lucia\Desktop\system-general-ai`
2. Ask: "Recuperá el plan del proyecto desde Engram"
3. The agent should call `mem_search` with topic key `system-general-ai/plan` and `personal` scope
4. Review `PLAN.md` and `TASKS.md` to confirm state matches
5. Resume from the first 🔄 or ⏳ stage

