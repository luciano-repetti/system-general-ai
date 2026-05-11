---
name: sdd-explore
description: Investigate an idea before committing to a change. Read codebase, compare approaches, surface risks. No files are created.
trigger: When the user describes a feature/refactor and wants options or feasibility assessment before proposing.
---

# sdd-explore

Read-only investigation. Output is a structured exploration saved to Engram.

## Steps

1. Restate the idea in your own words. Surface ambiguities.
2. Map the relevant code surface (files, modules, APIs touched).
3. List 2–3 viable approaches with trade-offs (effort, risk, maintenance cost).
4. Identify open questions the user should answer before proposing.
5. Flag risks: regressions, data migrations, security, performance.

## Save

`topic_key: sdd/{change-name}/explore`, scope `project`. Include:
- restated_idea
- code_surface (file paths)
- approaches (array of {name, pros, cons, effort})
- open_questions
- risks
- recommended_next_step (propose / clarify / abandon)

## Output (result contract)

```
status: complete | needs-input | aborted
executive_summary: 2-3 sentences
artifacts: [{type: explore, topic_key, observation_id}]
next_recommended: propose | clarify | abandon
risks: [array of strings]
skill_resolution: injected | fallback-* | none
```

Do NOT create files. Do NOT modify code. Investigation only.
