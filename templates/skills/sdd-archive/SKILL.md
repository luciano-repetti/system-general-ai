---
name: sdd-archive
description: Close a change. Persist final state, mark DAG complete. Cheap copy-and-close phase.
trigger: After verify reports overall pass. Final SDD phase.
---

# sdd-archive

Finalize the change. Read all artifacts, snapshot, mark closed.

## Reads

Required:
- `mem_search(query: "sdd/{change-name}/proposal")`
- `mem_search(query: "sdd/{change-name}/spec")`
- `mem_search(query: "sdd/{change-name}/design")` (if exists)
- `mem_search(query: "sdd/{change-name}/tasks")`
- `mem_search(query: "sdd/{change-name}/apply-progress")`
- `mem_search(query: "sdd/{change-name}/verify-report")`

## Steps

1. Confirm verify-report has `overall: pass`. If not, abort.
2. Compose an archive report covering: what changed, why, key decisions, lessons learned.
3. Save to Engram with `topic_key: sdd/{change-name}/archive-report`, scope `project`.
4. Update `sdd/{change-name}/state` with `status: archived`, `archived_at: <timestamp>`.

## Refuse

If `verify-report.overall` is not `pass`, refuse to archive. Return `status: blocked` with reason.

## Output

```
status: complete | blocked
executive_summary
artifacts: [{type: archive-report, topic_key, observation_id}]
change_summary: 1-2 sentences
```

This phase intentionally does the minimum work — copying summaries and toggling state. Use the `haiku` model.
