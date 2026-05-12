---
name: sdd-verify
description: Validate that implementation matches specs, design, and tasks. Reports CRITICAL / WARNING / SUGGESTION findings.
trigger: After apply reports complete. Required before archive.
---

# sdd-verify

Independent verification. Read-only validation against the contract.

## Reads

Required:
- `mem_search(query: "sdd/{change-name}/spec")` then `mem_get_observation`.
- `mem_search(query: "sdd/{change-name}/tasks")` then `mem_get_observation`.
- `mem_search(query: "sdd/{change-name}/apply-progress")` then `mem_get_observation`.

## Steps

1. For each requirement in spec: check the implementation honors it. Read code, run tests if available.
2. For each task in the checklist: confirm it was actually done (not just marked).
3. Run the full test suite. Capture pass/fail and failures.
4. Run lint/type-check if configured.
5. Check invariants from spec are not broken by the diff.
6. Classify findings:
   - **CRITICAL**: spec violation or test failure. Must fix before archive.
   - **WARNING**: best-practice violation, missing test, edge case unhandled.
   - **SUGGESTION**: nice-to-have, refactor opportunity.

## Save

`topic_key: sdd/{change-name}/verify-report`, scope `project`. Include:
- requirements_status (array of {id, pass: true/false, notes})
- tasks_status (array of {id, done_actually: true/false, notes})
- test_output (last run)
- findings (array of {severity, message, location})
- overall: pass | fail | conditional

## Output

```
status: complete
executive_summary
artifacts: [{type: verify-report, topic_key, observation_id}]
findings_count: {critical, warning, suggestion}
overall: pass | fail | conditional
next_recommended: archive | apply-fixes
```

## Enforcement (issue #262)

`overall: pass` is only valid if:
- All requirements pass
- Test suite is green
- No CRITICAL findings

If any of those is false, set `overall: fail` regardless of warnings/suggestions.
