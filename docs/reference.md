# Reference

Reference tables. No narrative.

## Permission profiles

Source: `templates/settings/{permissive,balanced,strict}.json`. The `defaultMode` controls Claude Code's behavior when a tool call doesn't match any list.

| Profile | `defaultMode` | Allow (key entries) | Ask (key entries) | Deny (all profiles) |
|---|---|---|---|---|
| `permissive` | `bypassPermissions` | Read, Grep, Glob, Edit, Write, Bash, WebFetch, WebSearch, Task*, Agent, Skill, ToolSearch, NotebookEdit, `mcp__*` | — (model self-regulates under bypass) | `rm -rf /*`, `rm -rf ~`, `rm -rf $HOME*`, `dd if=*`, `mkfs*`, `git push --force*`, `git push -f*`, `git reset --hard*`, `git clean -fdx*`, `git branch -D main*`, `git branch -D master*`, fork bomb |
| `balanced` | unset (Claude Code default) | Read, Grep, Glob, Task*, Skill, ToolSearch, narrow Bash (`git status*`, `git diff*`, `git log*`, `ls*`, `pwd`, `cat*`, `head*`, `tail*`, `which*`, `env`, `printenv*`), `mcp__plugin_engram_engram__mem_*` | Edit, Write, Bash (broad), WebFetch, WebSearch, Agent, EnterWorktree, ExitWorktree, NotebookEdit | same |
| `strict` | unset | Read, Grep, Glob, TaskGet, TaskList, ToolSearch, narrow Bash (`git status*`, `git diff*`, `git log*`, `pwd`, `ls`), only Engram read tools (`mem_search`, `mem_context`, `mem_get_observation`) | Edit, Write, Bash (broad), WebFetch, WebSearch, Agent, Task, TaskCreate, TaskUpdate, Skill, EnterWorktree, ExitWorktree, NotebookEdit, `mcp__*` | same |

`permissive` trades up-front prompts for trust + a deny list. `balanced` asks before mutating tools. `strict` asks for almost everything mutating, including all MCP servers.

## SDD behavior by target

| Target | Entry point | Behavior |
|---|---|---|
| Claude Code | `<instance>/CLAUDE.md` + skills | Coordinator-first SDD with delegation rules and phase skills. |
| Gemini CLI | project/global `GEMINI.md` + skills | Same SDD phase model adapted to Gemini global skills. |
| Codex | `~/.codex/AGENTS.md`, project `AGENTS.md`, `~/.codex/skills/` | Smallest useful SDD path. Direct answers, simple lookups, and trivial one-file changes do not force phase ceremony. Subagents are conditional on availability and value. |

Codex also receives `model_instructions_file` and `experimental_compact_prompt_file` entries for Engram in `~/.codex/config.toml`.

## SDD model assignments

Source: `templates/orchestrator.md` for Claude Code and the shared phase model.

| Phase | Model | Reason |
|---|---|---|
| `orchestrator` | opus | Coordinates, makes decisions |
| `sdd-explore` | sonnet | Reads code, structural |
| `sdd-propose` | opus | Architectural decisions |
| `sdd-spec` | sonnet | Structured writing |
| `sdd-design` | opus | Architecture decisions |
| `sdd-tasks` | sonnet | Mechanical breakdown |
| `sdd-apply` | sonnet | Implementation |
| `sdd-verify` | sonnet | Validation |
| `sdd-archive` | haiku | Copy and close |
| `shell-runner` | haiku | Execute and summarize |
| default delegation | sonnet | — |

If the assigned model isn't available, sub-agents substitute `sonnet` and continue.

Codex templates do not assign Claude model names. They express orchestration policy, Engram usage, phase order, and delegation rules in `templates/codex/orchestrator.md`.

## Skills shipped

11 skills, installed as directories:

- Claude Code: `<instance>/skills/<name>/SKILL.md`
- Gemini CLI: `~/.gemini/skills/<name>/SKILL.md`
- Codex: `~/.codex/skills/<name>/SKILL.md`

| Skill | Purpose |
|---|---|
| `sdd-init` | Initialize SDD context for a project (stack, test runner, conventions) |
| `sdd-explore` | Investigate an idea before committing — no files created |
| `sdd-propose` | Create a change proposal (intent, scope, approach) |
| `sdd-spec` | Write delta specs (requirements + scenarios) |
| `sdd-design` | Technical design with architecture decisions |
| `sdd-tasks` | Break a change into an implementation checklist |
| `sdd-apply` | Implement tasks, update apply-progress as work happens |
| `sdd-verify` | Validate implementation matches specs/design/tasks |
| `sdd-archive` | Close a change, persist final state |
| `shell-runner` | Run a command in a Haiku sub-agent and return only a summary |
| `compact-suggest` | Surface `/compact` when the conversation accumulates |

## Files sgai owns (manifest)

What `install`/`sync` writes and what `uninstall` removes. Source: `internal/claude/sync.go`, `internal/gemini/sync.go`, `internal/codex/sync.go`, `internal/cli/uninstall.go`.

| Artifact | Path | Install behavior | Uninstall behavior |
|---|---|---|---|
| Output style | `<instance>/output-styles/system-general-ai.md` | Overwrites with template | File removed |
| Global rules block | `<instance>/CLAUDE.md` between `<!-- BEGIN: system-general-ai/global-rules -->` markers | Merge in place; preserves user content outside markers | Marker block removed; rest of file untouched |
| Orchestrator block | `<instance>/CLAUDE.md` between `<!-- BEGIN: system-general-ai/sdd-orchestrator -->` markers | Same; absent when SDD off | Marker block removed |
| `outputStyle` key | `<instance>/settings.json` | Set to `system-general-ai` | Removed only if value still equals `system-general-ai` |
| `permissions` block | `<instance>/settings.json` | Deep-merged: union of `allow`/`deny`/`ask` (template-first dedup); user `defaultMode` wins | Removed only if shape matches one of our presets (heuristic on deny-list signature) |
| `$schema` key | `<instance>/settings.json` | Set to JSON schemastore URL | Left in place (conservative) |
| Skills | `<instance>/skills/<name>/SKILL.md` for each of the 11 names | Overwrites `SKILL.md` | Removes `SKILL.md`; removes the dir only if empty (preserves user-authored extras) |
| Engram MCP entry | `<instance>/settings.json` `mcpServers.engram` | Set to `{command, args:["serve"], env:{}}` | Removed only if `command` path contains the sgai bin dir |
| Engram binary | `~/.local/bin/engram` (POSIX) or `%LOCALAPPDATA%\system-general-ai\bin\engram.exe` (Windows) | Downloaded by `install` | Kept unless `--remove-engram` is passed |

## Codex files sgai owns

| Artifact | Path | Install behavior | Uninstall behavior |
|---|---|---|---|
| Global rules | `~/.codex/AGENTS.md` between `<!-- BEGIN: system-general-ai/... -->` markers | Merge in place; preserves user content outside markers | Managed marker blocks removed |
| Project rules | `./AGENTS.md` between `<!-- BEGIN: system-general-ai/project-rules -->` markers | Merge in current project | Managed marker block removed |
| Engram MCP | `~/.codex/config.toml` `[mcp_servers.engram]` | Upserts managed TOML block | Managed block removed |
| Engram model instructions | `~/.codex/engram-instructions.md` | Overwrites with template | File removed |
| Engram compact prompt | `~/.codex/engram-compact-prompt.md` | Overwrites with template | File removed |
| Skills | `~/.codex/skills/<name>/SKILL.md` | Overwrites managed `SKILL.md` files | Removes managed `SKILL.md`; preserves user extras in the directory |
| Backups | `~/.codex/backups/system-general-ai-*` | Created before modifying Codex config files | Left in place |

Run Codex install/sync from the project directory that should receive `./AGENTS.md`. The global files are always written under `CODEX_HOME` when set, otherwise `~/.codex`.

## Engram MCP entry

Written by `syncEngramMCP` (`internal/claude/sync.go:171`). Shape:

```json
{
  "mcpServers": {
    "engram": {
      "command": "/abs/path/to/engram",
      "args": ["serve"],
      "env": {}
    }
  }
}
```

Other entries under `mcpServers` are preserved.

## Codex Engram MCP entry

Written to `~/.codex/config.toml`:

```toml
[mcp_servers.engram]
command = "/abs/path/to/engram"
args = ["mcp", "--tools=agent"]
enabled = true
startup_timeout_sec = 20
tool_timeout_sec = 60
```

Top-level Codex instruction files are also set:

```toml
model_instructions_file = "~/.codex/engram-instructions.md"
experimental_compact_prompt_file = "~/.codex/engram-compact-prompt.md"
```

## Multi-Claude detection rules

Source: `internal/claude/detect.go:61`.

Order of precedence:

1. **`CLAUDE_CONFIG_DIR`** — if set and non-empty, this is treated as an EXCLUSIVE override. Only that directory is configured. The default + autodiscovery below are skipped.
2. **`$HOME/.claude`** — the default instance.
3. **`$HOME/.claude-work*`** — glob-based autodiscovery for multi-credential setups.

Duplicates (by absolute path) are deduped. Non-directory paths are ignored. Result is sorted by path for stable output.

