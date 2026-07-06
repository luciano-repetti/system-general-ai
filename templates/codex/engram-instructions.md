# Engram Persistent Memory Protocol

Engram is the persistent memory MCP. Use it for durable project knowledge, not as a transcript store. Use every Engram tool exposed in the session; if a needed tool is not visible, use tool discovery / ToolSearch with the exact tool name before treating it as unavailable.

## Tool Map

- Context/read: `mem_current_project`, `mem_context`, `mem_search`, `mem_get_observation`.
- Save/update: `mem_save`, `mem_update`, `mem_suggest_topic_key`, `mem_save_prompt`, `mem_capture_passive`.
- Session lifecycle: `mem_session_start`, `mem_session_summary`, `mem_session_end`.
- Conflict/relation handling: `mem_judge`, `mem_compare`.
- Diagnostics/admin: `mem_doctor`, `mem_stats`, `mem_timeline`, `mem_delete`, `mem_merge_projects`.

Deferred tools may require discovery. Search exact names such as `mem_judge`, `mem_update`, `mem_session_summary`, or `mem_doctor`; broad searches can miss tools.

Use destructive/admin tools such as `mem_delete` or `mem_merge_projects` only when explicitly requested or clearly safe.

## Search Budget

Do not call Engram by reflex.

- Direct answer, explanation, one-off lookup, or trivial one-file mechanical edit: skip Engram unless the user asks about past work.
- Small repo task: run one focused memory pass before editing: `mem_current_project` if project is uncertain, then one combined `mem_search` query for relevant decisions, bugs, commands, and SDD state. Add `mem_context` only when recent-session context matters.
- Medium/large task or SDD continuation: run `mem_context` plus targeted `mem_search` queries for `sdd-init/{project}`, active `sdd/{change}/...` artifacts, decisions, bugs, commands, and environment traps. Cache what you read; do not repeat the same search every turn.
- If a search result affects work and is truncated, call `mem_get_observation` for the full content before relying on it.

If Engram is unavailable, say so briefly and continue with local repo context; do not pretend memory was checked.

## Save Rules

Save only durable knowledge that would help a future agent avoid rediscovery or preserve intent:

- verified bug root cause or fix;
- architecture/design decision and trade-off;
- business rule, domain invariant, or integration contract;
- non-obvious codebase discovery;
- command, setup, or environment trap that took real effort;
- user preference or project convention likely to apply again;
- SDD phase artifact, apply progress, verify report, or archive.

Do not save:

- trivial answers, routine file lookups, or obvious facts;
- raw logs, full command output, secrets, credentials, tokens, or duplicate file contents;
- transient git status, commit hash, push result, or release note unless it changes future workflow;
- passing test/build output unless it is verification evidence for a saved decision/fix;
- duplicate memories; update/upsert instead.

Use this compact content shape:

```text
What: one sentence.
Why: reason or impact.
Where: files, modules, commands, or systems.
Evidence: test, build, observation, or reproduction.
Learned: gotcha, invariant, or trap.
Next: follow-up, if any.
```

## Topic Keys and Updates

Prefer project-scoped stable topic keys:

- `{project}/business-logic`
- `{project}/architecture`
- `{project}/decisions`
- `{project}/bugs`
- `{project}/commands`
- `{project}/environment`
- `{project}/testing`
- `sdd-init/{project}`
- `sdd/{change}/explore|proposal|spec|design|tasks|apply-progress|verify-report|archive-report|state`

Use `mem_suggest_topic_key` when the stable key is unclear. Use `mem_update` when you have the exact observation ID. Use `mem_save` with the same `topic_key` for evolving project state so Engram can upsert instead of creating duplicates.

## Conflict Handling

If `mem_save` returns `judgment_required: true`, inspect every candidate and call `mem_judge` once per candidate using that candidate's `judgment_id`.

Ask the user before judging only when confidence is low, or when the likely relation is `supersedes` / `conflicts_with` for architecture, policy, or decision memories. Otherwise resolve silently with `related`, `compatible`, `scoped`, or `not_conflict`.

Use `mem_compare` only when explicitly comparing two existing memories or when a conflict requires a durable relation outside the `mem_save` flow.

## Session Close and Compaction

For substantial work, call `mem_session_summary` before final response with goal, decisions, discoveries, completed work, verification, next steps, and relevant files. Skip session summaries for direct answers and trivial work.

After compaction, first persist the compacted summary with `mem_session_summary`, then recover context with `mem_context` and targeted `mem_search` before continuing.
