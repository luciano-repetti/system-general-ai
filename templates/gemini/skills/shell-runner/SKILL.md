---
name: shell-runner
description: Execute a shell command, read its output, and return a summary preserving errors, exit codes, and key data. Used for commands with anticipated large output to keep the orchestrator's context clean.
trigger: When a shell command will produce more than 5000 tokens of output (test runs, builds, log dumps, large diffs, file listings).
---

# shell-runner

Execute, read, summarize. The full output stays in this sub-agent's context. Only the summary returns to the orchestrator.

## Inputs (from the orchestrator)

- `command` — the exact shell command to execute (string)
- `purpose` — one sentence describing why this command is being run, so the summary stays relevant
- `cwd` (optional) — working directory; defaults to project root

## Steps

1. Execute the command via Bash. Capture stdout, stderr, exit code.
2. Identify in the output:
   - Exit code (0 = success; non-zero = failure with reason)
   - Real errors (stderr lines that are actual failures, not warnings)
   - For test runners: pass/fail counts, failing test names with their first assertion message
   - For builds: success or first compilation error with file:line
   - For diffs: file count, lines added/removed, 2–3 highlight changes
   - Key data the orchestrator needs given the stated purpose
3. Compose a summary:
   - State exit code first (PASS / FAIL with reason).
   - List concrete errors with location (file:line when available).
   - Omit verbose noise (timestamps, progress bars, repeated lines, color codes).

## Output (return to orchestrator)

```
exit_code: <number>
status: success | failure
summary: <2–10 sentences>
errors: <bullet list — empty if exit_code is 0>
key_data: <relevant excerpts only, no raw dumps>
```

## Limits

- Hard cap: 500 tokens in the returned summary.
- If the command output is binary or unreadable, return `status: error` with the reason.
- Refuse destructive commands (`rm -rf /`, `git reset --hard`, force push, drop database, etc).

## Model

Assigned to `haiku` — fast, cheap, sufficient for read-and-summarize.
