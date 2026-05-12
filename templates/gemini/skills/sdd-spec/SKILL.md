---
name: sdd-spec
description: Write specifications with requirements and scenarios. Delta specs for changes (what changes from current behavior).
trigger: After proposal is approved. Required input for tasks generation.
---

# sdd-spec

Write the contract. What must be true after this change ships.

## Reads

Required: `mem_search(query: "sdd/{change-name}/proposal")` then `mem_get_observation`.

## Steps

1. Translate acceptance criteria into testable requirements.
2. Write requirements as **delta specs** — describe the change vs current behavior, not the full system.
3. For each requirement, list scenarios (Given / When / Then style is fine, but plain text is acceptable).
4. Mark requirements as functional vs non-functional (perf, security, observability).
5. Declare invariants that must NOT change.

## Save

`topic_key: sdd/{change-name}/spec`, scope `project`. Include:
- requirements (array of {id, type, description, scenarios})
- invariants
- non_functional_requirements

## Output

```
status: complete | blocked
executive_summary
artifacts: [{type: spec, topic_key, observation_id}]
next_recommended: design | tasks (if design optional)
```
