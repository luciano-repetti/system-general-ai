# Caching and Memory

Two related concerns: (1) keeping Anthropic's automatic prompt caching effective, (2) using Engram so the conversation history can be discarded without losing context. `system-general-ai` is designed around both.

## Automatic prompt caching

Claude Code uses Anthropic's prompt caching automatically. The cache key is the stable prefix of each request — Anthropic's base system prompt, the tool schemas, the output style, and `CLAUDE.md`. After the first request in a 5-minute window, the cached prefix is billed at ~10% of nominal.

What this means in practice: as long as sgai's templates don't change between turns, the persona + global rules + orchestrator block all get cached. Per-turn overhead drops from ~3,000 tokens nominal to ~300 effective. See `docs/token-comparison.md` for the measured numbers (~11,862 bytes always-loaded, ~2,966 token estimate, ~300 after caching).

## Why sgai's design is cache-friendly

| Mechanism | Where | Effect |
|---|---|---|
| Marker-delimited merges | `mergeMarkedSection` (`internal/claude/sync.go:206`) | Re-running `sync` produces byte-identical output when nothing changed → cache stays warm |
| Deep-merge of `permissions` | `mergePermissions` (`internal/claude/sync.go:309`) | User customizations don't get clobbered, so the user has no reason to manually re-edit `settings.json` mid-session |
| Stable list ordering | `unionStringList` (`internal/claude/sync.go:351`) | Template-first dedup → idempotent output → stable cache |
| `Edit` over `Write` (rule) | `templates/CLAUDE.md.global.tmpl:73` | Tool result is a small confirmation, not a full file dump |
| Skills loaded on demand | Claude Code parser | Skill bodies are NOT in the always-loaded prefix; only the directory listing is |

## Practical tips

- Don't have the agent rewrite `CLAUDE.md` or `settings.json` mid-session unless the user asked. Stable system content = cache hits.
- Prefer `Edit` over `Write` for existing files. `Write` returns full content; `Edit` returns the diff.
- Run `sgai install` and `sgai sync` outside an active Claude Code session when possible. Restarting Claude Code re-warms the cache from clean state.

## Engram persistent memory

Engram (an MCP server) gives Claude Code memory that survives session boundaries and `/compact`. The protocol sgai installs (`templates/CLAUDE.md.global.tmpl:49`) requires the agent to **save proactively** — not wait to be asked.

| When to save (`mem_save`) | Examples |
|---|---|
| Architecture / design decisions | "Chose CQRS over service layer because…" |
| Bug fixes | Include root cause |
| Patterns established | Naming conventions, file layout |
| User preferences / constraints | "Do not run builds after changes" |
| Non-obvious discoveries | "Function X has a hidden coupling to Y" |

| When to search (`mem_search` then `mem_get_observation`) | |
|---|---|
| User says "remember", "what did we do", "acordate", references prior work |
| Starting a task that may have been done before |
| First user message references a topic with no in-session context |

`mem_search` returns truncated results — always follow up with `mem_get_observation` for the full content.

## Topic key convention

Stable keys so updates upsert in place rather than create duplicates:

| Artifact | Topic key |
|---|---|
| Project context | `sdd-init/{project}` |
| Exploration | `sdd/{change}/explore` |
| Proposal | `sdd/{change}/proposal` |
| Spec / Design / Tasks | `sdd/{change}/{spec\|design\|tasks}` |
| Apply progress | `sdd/{change}/apply-progress` |
| Verify report | `sdd/{change}/verify-report` |
| Archive report | `sdd/{change}/archive-report` |

Source: `templates/orchestrator.md:60`.

## Scope

- `project` (default) — visible only to the current project
- `personal` — cross-project, follows the user (preferences, recurring constraints)

## The `shell-runner` sub-agent pattern

`templates/skills/shell-runner/SKILL.md` defines a Haiku sub-agent that takes a command, executes it, and returns a 2–10 sentence summary capped at 500 tokens. The full output stays inside the sub-agent's context — the main orchestrator never sees it.

Triggered for any command with anticipated output >5,000 tokens: test runs, builds, log dumps, large diffs, file listings (`templates/orchestrator.md:36`). Short stateful commands (`git status`, `ls`, `pwd`) execute directly.

Sub-agents are billed separately. A test run that would cost ~30k Sonnet input tokens in the main context costs the same ~30k as Haiku input, ~10× cheaper, AND the main context never grows.

## The `compact-suggest` skill

`templates/skills/compact-suggest/SKILL.md` is a lightweight Haiku check. The agent cannot run `/compact` itself — only the user can. This skill detects accumulation (>20 turns, >50,000 tokens of tool results, 3+ files edited with large diffs, 3+ Engram saves this session) and appends a non-intrusive single-line suggestion. Capped at one suggestion every 10 turns. Skipped mid-task.

The premise: if Engram has the important decisions, the chat history is disposable.

## Practical workflow

1. Start work. Persona + global rules + orchestrator are cached after turn 1.
2. Make decisions. Agent saves them to Engram automatically.
3. Run a long test. Orchestrator delegates to `shell-runner` (Haiku); summary returns; main context stays small.
4. Conversation grows. `compact-suggest` mentions `/compact` once.
5. User runs `/compact` or starts a new session. Engram restores the relevant context via `mem_search`.

