<!-- BEGIN: system-general-ai/project-rules -->
# system-general-ai Project Rules

This project uses system-general-ai with Codex.

## Defaults

- SDD is always on for non-trivial coding work, using the smallest useful path.
- Engram is always on for durable project memory.
- Subagents should be used only when they reduce latency, isolate independent work, or keep large context out of the main thread.
- Keep edits scoped to the requested behavior and the repo's existing patterns.

## Task Start

Before changing code:

1. Search Engram for project memory and active SDD state.
2. Inspect only the files needed to classify the task and verify assumptions.
3. Decide whether this is trivial, small, medium, or large.
4. Use the matching SDD path from the global rules.

## Project Memory

At the start of meaningful work, search Engram for this project before touching files. Save business logic, decisions, difficult discoveries, and verified bug fixes as they appear.

Save especially:

- domain entities and invariants;
- workflow rules;
- integration contracts;
- error conditions and recovery behavior;
- commands that are known to work;
- commands or setups that failed and why.

## Local Artifacts

Do not rely on conversation history as the only source of state. Persist SDD state and valuable project knowledge in Engram.

## Closeout

Before saying work is complete, verify the change or state why verification could not run. Save durable findings to Engram first.
<!-- END: system-general-ai/project-rules -->
