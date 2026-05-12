---
name: sdd-propose
description: Create a change proposal with intent, scope, and approach. Architectural decision happens here.
trigger: After exploration is done (or skipped intentionally) and the user wants to commit to an approach.
---

# sdd-propose

Architectural commitment. Decide the approach and scope.

## Reads (optional)

If exploration exists, read it first:
1. `mem_search(query: "sdd/{change-name}/explore")`
2. `mem_get_observation(id: ...)` for full content

## Steps

1. Define the goal in one sentence.
2. State scope: in-scope, out-of-scope.
3. Choose ONE approach (not a menu — a decision). Justify with 2–3 sentences.
4. Identify components/files affected.
5. List dependencies (other features, libraries, infra).
6. Define acceptance criteria — testable and concrete.
7. Estimate effort buckets (S / M / L / XL) — single number.

## Save

`topic_key: sdd/{change-name}/proposal`, scope `project`. Include:
- goal
- scope_in, scope_out
- chosen_approach + rationale
- affected_components
- dependencies
- acceptance_criteria
- effort

## Output

```
status: complete | blocked
executive_summary
artifacts: [{type: proposal, topic_key, observation_id}]
next_recommended: spec | design | tasks
risks
```

If a critical question is unanswered, return `status: blocked` with the question.
