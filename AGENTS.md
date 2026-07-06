<!-- BEGIN: system-general-ai/project-rules -->
# system-general-ai Project Rules

This project uses system-general-ai with Codex. Global rules define SDD, Engram, delegation, and output style; this file adds repo-local scope.

## Defaults

- Use the smallest SDD path that produces evidence.
- Keep edits scoped to requested behavior and existing repo patterns.
- Treat Engram as durable memory, not a chat transcript.
- Run Codex from the repo root when possible so Engram project binding is correct.

## Task Start

Before changing code for non-trivial work:

1. Use the global Engram budget: skip trivial work, one focused search for small work, targeted phase searches for medium/large work.
2. Inspect only files needed to classify the task and verify assumptions.
3. Decide whether this is trivial, small, medium, or large.
4. Use the matching SDD path from the global rules.

## Project Memory

Save only durable project knowledge: business logic, architecture decisions, difficult discoveries, verified bug fixes, environment traps, and commands that are known to work.

Do not save routine commits, pushes, obvious facts, raw outputs, or duplicate memories. Prefer stable topic keys and `mem_update`/upsert when updating existing knowledge.

## Local Artifacts

Do not rely on conversation history as the only source of state. Persist SDD state and valuable project knowledge in Engram using the topic keys from the global rules.

## Closeout

Before saying substantial work is complete, verify the change or state why verification could not run. Save durable findings to Engram first; skip memory for trivial/direct answers when nothing durable was learned.
<!-- END: system-general-ai/project-rules -->
