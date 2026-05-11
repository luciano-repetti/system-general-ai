---
name: sdd-design
description: Technical design document with architecture decisions and approach details. Optional but recommended for non-trivial changes.
trigger: After proposal. Skip for trivial changes (one-file edits, copy fixes).
---

# sdd-design

Architecture decision record + technical approach. Skip if change is trivial.

## Reads

Required: `mem_search(query: "sdd/{change-name}/proposal")` then `mem_get_observation`.

## Steps

1. Diagram the change in plain text or simple ASCII (no need for tools).
2. List architectural decisions made — each as a short ADR-style entry: context, decision, consequences.
3. Define interfaces / contracts touched (function signatures, API endpoints, DB schema).
4. Note migrations needed (data, config, dependencies).
5. Define rollback strategy if the change goes wrong.
6. Performance / security considerations if applicable.

## Save

`topic_key: sdd/{change-name}/design`, scope `project`. Include:
- diagram (text)
- decisions (array of {context, decision, consequences})
- interfaces
- migrations
- rollback_strategy
- performance_notes
- security_notes

## Output

```
status: complete | blocked
executive_summary
artifacts: [{type: design, topic_key, observation_id}]
next_recommended: tasks
```
