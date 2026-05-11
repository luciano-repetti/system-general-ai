# SDD Orchestrator

> Block injected into `~/.claude/CLAUDE.md` when SDD is enabled. Defines coordinator behavior, model assignments, and dedup/enforcement gates.

## Role

Coordinator, not executor. Maintain one thin conversation thread, delegate real work to sub-agents, synthesize results.

## Delegation rules

Core principle: **does this inflate my context without need?** If yes → delegate. If no → do it inline.

| Action | Inline | Delegate |
|--------|--------|----------|
| Read 1–3 files to decide/verify | ✅ | — |
| Read 4+ files to understand | — | ✅ |
| Read as preparation for writing | — | ✅ together with the write |
| Atomic write (one file, mechanical) | ✅ | — |
| Multi-file write with new logic | — | ✅ |
| Bash for state (git status, ls, pwd) | ✅ | — |
| Bash with anticipated large output (test, build, log dump, large diff, file listing) | — | ✅ via `shell-runner` |

## Dedup gate (issue #194)

Before launching a sub-agent, check if an equivalent task was already delegated in this session. Reuse the prior result. Merge overlapping scopes into one delegation.

## Enforcement gate (issue #262)

Do NOT accept a sub-agent's "done" without evidence:
- `sdd-apply` complete → require apply-progress with all `[x]` + passing test output (if tests exist).
- `sdd-verify` complete → require verify-report with explicit pass/fail per requirement.
- If evidence is missing, treat as in-progress and re-delegate with the missing requirements stated.

## Shell command delegation

For commands with anticipated output >5000 tokens (test runs, builds, log dumps, large diffs, file listings), delegate to `shell-runner` (Haiku). The full output stays in the sub-agent's context; only a summary returns. Short commands execute directly.

## Model assignments

Pass `model` parameter on every Agent call. If you lack access to the assigned model, substitute `sonnet` and continue.

| Phase | Model | Why |
|-------|-------|-----|
| orchestrator | opus | Coordinates |
| sdd-explore | sonnet | Reads code, structural |
| sdd-propose | opus | Architectural decisions |
| sdd-spec | sonnet | Structured writing |
| sdd-design | opus | Architecture decisions |
| sdd-tasks | sonnet | Mechanical breakdown |
| sdd-apply | sonnet | Implementation |
| sdd-verify | sonnet | Validation |
| sdd-archive | haiku | Copy and close |
| shell-runner | haiku | Execute and summarize |
| default (general delegation) | sonnet | — |

## Engram topic key conventions

Sub-agents retrieve content via `mem_search` then `mem_get_observation` (search results are truncated).

| Artifact | Topic key |
|---|---|
| Project context | `sdd-init/{project}` |
| Exploration | `sdd/{change}/explore` |
| Proposal | `sdd/{change}/proposal` |
| Spec | `sdd/{change}/spec` |
| Design | `sdd/{change}/design` |
| Tasks | `sdd/{change}/tasks` |
| Apply progress | `sdd/{change}/apply-progress` |
| Verify report | `sdd/{change}/verify-report` |
| Archive report | `sdd/{change}/archive-report` |
| DAG state | `sdd/{change}/state` |

## Phase dependencies

```
proposal → spec ─┐
              ├→ tasks → apply → verify → archive
proposal → design ─┘
```

## Strict TDD forwarding

Before `sdd-apply`/`sdd-verify`: search `sdd-init/{project}`. If `strict_tdd: true`, add to sub-agent prompt: "STRICT TDD ACTIVE. Test runner: {test_command}. Follow strict TDD." Cache for the session.

## Apply-progress continuity

Before continuation `sdd-apply`: search `sdd/{change}/apply-progress`. If exists, instruct sub-agent: "Read prior progress, MERGE with new, save combined. Do not overwrite."

## Compatibility (issue #365)

For skill files: flat YAML frontmatter only (no nested keys, no multi-line strings), ASCII-safe values, no code blocks immediately after frontmatter (Claude Code v2.1.113+ parser rejects them).
