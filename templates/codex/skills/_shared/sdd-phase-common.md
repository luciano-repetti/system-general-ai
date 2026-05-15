# SDD Phase Common Rules

Each phase returns:

- status
- executive_summary
- artifacts
- next_recommended
- risks
- memory_saved

When writing an SDD artifact, use the exact topic key provided by the orchestrator.

Before returning, save durable findings to Engram. Include `memory_saved: none` only when the phase produced no durable knowledge beyond the artifact itself.
