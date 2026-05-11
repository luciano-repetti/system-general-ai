---
name: sdd-apply
description: Implement tasks from the change. Execute the breakdown. Updates apply-progress as work happens.
trigger: After tasks are saved. Re-runs continue from previous progress (read-merge-write).
---

# sdd-apply

Execute the implementation. Real code changes happen here.

## Reads

Required:
- `mem_search(query: "sdd/{change-name}/tasks")` then `mem_get_observation`.
- `mem_search(query: "sdd/{change-name}/spec")` then `mem_get_observation`.
- `mem_search(query: "sdd/{change-name}/design")` (if exists).
- `mem_search(query: "sdd/{change-name}/apply-progress")` — if exists, READ FIRST, MERGE with new progress before saving (do NOT overwrite).

## Strict TDD mode

If active (forwarded by orchestrator):
- Write the failing test BEFORE the implementation.
- Run the test, confirm it fails for the right reason.
- Implement. Re-run. Confirm it passes.
- Refactor only when green. Re-run after refactor.
- One requirement → one test → one implementation cycle.

If NOT active: implementation can proceed without writing tests first, but verify must still pass.

## Steps

1. Pick the first incomplete task from the checklist.
2. Implement the change. Edit existing files; create only when truly needed.
3. Run the relevant verification command (test, build, lint).
4. Mark the task as `[x]` in the apply-progress.
5. Save apply-progress to Engram (topic_key: `sdd/{change-name}/apply-progress`) — MERGED with prior content.
6. Repeat until batch is done.

## Stop conditions

- All tasks complete → return `status: complete`.
- Verification fails on a task → fix or report `status: blocked` with the failure.
- Hit a previously unconsidered case → return `status: needs-design` with the case.

## Save (apply-progress)

After each task or batch:

```markdown
- [x] T1. Done
- [x] T2. Done
- [ ] T3. In progress — blocked on X
```

Plus structured fields: `tasks_completed`, `tasks_remaining`, `last_run_test_output`, `blockers`.

## Output

```
status: complete | partial | blocked | needs-design
executive_summary
artifacts: [{type: apply-progress, topic_key, observation_id}]
tasks_completed: N of M
test_output: <last run summary>
next_recommended: verify | continue-apply | redesign
```

## Enforcement (issue #262)

Never report `status: complete` without evidence. The orchestrator will reject it. Required:
- Apply-progress with all tasks `[x]`
- Test command output proving suite passes (if tests exist)
