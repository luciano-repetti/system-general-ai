---
name: compact-suggest
description: Detect when conversation history is accumulating and suggest the user run /compact to consolidate the session. Does not compact anything itself.
trigger: When 20+ turns have passed, 50000+ tokens of tool results have accumulated, or 3+ files have been edited with large diffs.
---

# compact-suggest

Lightweight session-health check. The agent cannot invoke `/compact` (only the user can). This skill detects accumulation and surfaces a single, non-intrusive suggestion.

## When to invoke

Trigger if any of these is true:
- Conversation has >20 turns
- Tool results have accumulated >50000 tokens
- 3+ files have been edited with diffs >200 lines each
- Engram has stored >3 decisions/discoveries in this session (history is no longer needed for them — they are recoverable)

## What to do

Output a brief suggestion. Do NOT make it the focus of the reply — append it after the user's actual answer.

In Spanish:
> *"La conversación se está acumulando. Si querés bajar el contexto, podés correr `/compact` — lo importante ya está en Engram, no se pierde."*

In English:
> *"Conversation is accumulating. To reduce context, run `/compact` — the important decisions are already in Engram, nothing is lost."*

## Limits

- Suggest at most ONCE every 10 turns. Don't pester.
- If the user ignored a previous suggestion within the last 5 turns, do NOT repeat.
- Never auto-execute. Only suggest.
- Skip the suggestion if a focused task is in progress and `/compact` would interrupt mid-flow.

## Model

Assigned to `haiku` — quick check, no architecture.
