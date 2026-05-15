# Engram Persistent Memory Protocol

Engram is mandatory and always active. Use it to preserve project knowledge across sessions, compaction, and agent delegation.

## First Action For Meaningful Work

Before touching code for a non-trivial task:

1. Search for project memory.
2. Search for business logic and prior decisions.
3. Search for active SDD state if the task may continue previous work.
4. Retrieve full observations with `mem_get_observation` before relying on truncated search results.

## Search First

Before meaningful work, search Engram for relevant project context:

- business logic
- prior decisions
- architecture
- known bugs
- environment traps
- test/build commands
- active SDD state

If the first search result is truncated, call `mem_get_observation` for the full content before relying on it.

## Save Immediately

Save to Engram without being asked after any of these:

- business rule or domain invariant discovered
- architecture/design decision made
- bug root cause verified
- difficult or time-consuming discovery made
- multiple failed attempts before finding the right path
- environment/configuration trap found
- useful command or workflow established
- SDD phase completed
- user preference or constraint learned

Use compact, searchable content:

- What changed or was learned
- Why it matters
- Where it applies
- Evidence or verification
- Next implication, if any

Do not save secrets, credentials, raw logs, full command outputs, duplicate file contents, or obvious facts.

## Save Format

Use this shape:

```text
What: one sentence.
Why: reason or user/business impact.
Where: files, modules, commands, or systems.
Evidence: tests, build, observation, or reproduction.
Learned: gotchas, invariants, or traps.
Next: concrete follow-up, if any.
```

## Suggested Topic Keys

- `{project}/business-logic`
- `{project}/architecture`
- `{project}/decisions`
- `{project}/bugs`
- `{project}/commands`
- `{project}/environment`
- `{project}/testing`
- `{project}/sdd/current`
- `{project}/sdd/archive/{change}`

For SDD phase artifacts, use the orchestrator's exact `sdd/{change}/...` topic keys.

## Subagents

When launching subagents, pass relevant Engram topic keys. Subagents should save important findings before returning, because they have the freshest detail.

## Session Close

Before ending substantial work, save a session summary with:

- Goal
- Decisions
- Discoveries
- Completed work
- Verification
- Next steps
- Relevant files

## After Compaction

If the context was compacted, first persist the compacted summary to Engram, then recover project context from Engram before continuing.

## Anti-Noise Rule

If a memory would not help a future agent avoid rediscovery or preserve business/technical intent, do not save it.
