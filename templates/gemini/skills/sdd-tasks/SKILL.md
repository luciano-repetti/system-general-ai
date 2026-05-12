---
name: sdd-tasks
description: Break a change into a checklist of implementation tasks. Mechanical breakdown of spec + design.
trigger: After spec (and design if applicable) are saved. Required input for apply.
---

# sdd-tasks

Translate spec + design into an actionable checklist.

## Reads

Required:
- `mem_search(query: "sdd/{change-name}/spec")` then `mem_get_observation`.
- `mem_search(query: "sdd/{change-name}/design")` (optional but read if exists).

## Steps

1. Decompose the work into atomic tasks.
2. Each task: one outcome, testable, ≤ 2 hours of work.
3. Order tasks by dependency. Surface parallel branches.
4. For each task, list: file(s) touched, expected change, verification step.
5. Include test tasks explicitly when strict TDD is active (one test task per requirement).

## Save

`topic_key: sdd/{change-name}/tasks`, scope `project`. Format as Markdown checklist:

```markdown
- [ ] T1. Short description
      - Files: path/to/file.ext
      - Verify: command or expected diff
- [ ] T2. ...
```

Total line of array of task objects with `id`, `description`, `files`, `verify`, `status: pending`.

## Output

```
status: complete | blocked
executive_summary
artifacts: [{type: tasks, topic_key, observation_id}]
task_count: N
next_recommended: apply
```
