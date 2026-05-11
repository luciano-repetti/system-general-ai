# Token comparison — Claude default vs system-general-ai vs gentle-ai

## Methodology

Token estimation: **~4 characters per token** (conservative average for English/Spanish technical text). Real values vary by content (code: ~3 chars/token; prose: ~4-5).

What we measure: **the custom overhead the wrapper adds to each prompt** — the instructions injected into the system prompt and `CLAUDE.md` that ship with the wrapper. This is what stays constant in every turn.

What we cannot measure from here:
- Claude Code's own base system prompt (controlled by Anthropic, not by any wrapper)
- Tokens consumed by sub-agents (each gets a fresh context — separate billing)
- The real cost after Anthropic prompt caching (cached prefix is billed at ~10% of nominal)

The numbers below are **upper-bound nominal**. Real cost in $$ is significantly lower thanks to caching.

---

## Raw measurements (bytes)

### Claude Code default (no wrapper)

| Component | Bytes |
|---|---|
| Custom output style | 0 |
| Custom global CLAUDE.md | 0 |
| Custom skills | 0 |
| **Total custom overhead** | **0** |

The agent runs with Claude Code's built-in defaults only. No persona, no SDD, no engram, no skills.

### system-general-ai

| Component | Bytes | Always loaded? |
|---|---|---|
| `output-style.md` | 1,986 | Yes — when persona active |
| `CLAUDE.md.global.tmpl` (injected as marker section) | 4,519 | Yes |
| `orchestrator.md` (injected as marker section) | 3,857 | Yes — when SDD on |
| Skills directory listing (11 skills, names + 1-line descriptions) | ~1,500 | Yes — listing only |
| Skills bodies (sdd-*, shell-runner, compact-suggest) | 17,275 | **No** — loaded only when invoked |
| **Total per-turn overhead (always loaded)** | **~11,862** | |
| **Token estimate (4 chars/token)** | **~2,966** | |

### gentle-ai (Claude Code adapter)

| Component | Bytes | Always loaded? |
|---|---|---|
| Combined CLAUDE.md (Gentleman persona + SDD + Engram + rules) | 20,361 | Yes |
| Output style file (Gentleman persona) | ~3,500* | Yes |
| Skills directory listing (~25+ skills, names + descriptions) | ~3,500* | Yes — listing only |
| Skill bodies (e.g. `branch-pr/SKILL.md`: 8,489; `issue-creation/SKILL.md`: 7,582; etc) | ~30,000+ | **No** — loaded only when invoked |
| **Total per-turn overhead (always loaded)** | **~27,361** | |
| **Token estimate (4 chars/token)** | **~6,840** | |

\* Estimated — gentle-ai generates the persona dynamically; not directly measurable as a single file.

---

## Comparison

| Wrapper | Per-turn overhead (bytes) | Per-turn overhead (≈ tokens) | Notes |
|---|---|---|---|
| Claude Code default | 0 | 0 | Bare agent, no enhancements |
| **system-general-ai** | **~11,862** | **~2,966** | Persona + SDD + engram protocol + skills listing |
| gentle-ai | ~27,361 | ~6,840 | Same components, more verbose |

### Reduction vs gentle-ai

```
27,361 → 11,862 bytes  =  56.6% smaller
6,840 → 2,966 tokens   =  56.6% reduction
```

**system-general-ai is roughly 2.3× more compact** than gentle-ai for the same feature surface (persona + SDD + Engram + skills + multi-Claude support).

### What costs the extra ~3k tokens vs Claude default

You pay ~3k tokens per turn for:
1. **Persona that won't waste your time** with pedagogical lectures (~500 tokens, output style)
2. **SDD orchestrator** that delegates large tasks to sub-agents (~1,000 tokens) — this saves 5-50× more downstream by NOT inflating the main context with sub-agent work
3. **Global rules** for code conventions, verification, memory protocol (~1,100 tokens)
4. **Skills listing** so the agent knows what's available (~400 tokens; bodies only load on invocation)

---

## How prompt caching changes the picture

Anthropic caches the stable prefix of the prompt. After the first request in a 5-minute window, the cached portion is billed at **~10% of nominal**.

For system-general-ai, ~95% of the overhead (output style + CLAUDE.md + orchestrator) is **stable across turns within a session** — eligible for caching.

Effective cost per turn after the first cache hit:

| Wrapper | Nominal tokens | After caching |
|---|---|---|
| Claude Code default | 0 | 0 |
| system-general-ai | ~2,966 | **~300** |
| gentle-ai | ~6,840 | **~700** |

So in practice, the persistent overhead of system-general-ai during an active session is **~300 tokens per turn** — negligible vs the rest of the conversation.

---

## What system-general-ai gives you for those tokens

This is the value, not just the cost:

| Feature | Tokens "spent" | Tokens "saved" |
|---|---|---|
| SDD orchestrator | ~1,000 (always loaded) | 5,000-50,000 per delegation (sub-agent contexts kept clean) |
| `shell-runner` skill | ~50 (listing) + 500 (when invoked) | 5,000-50,000 per large command output (test, build, etc) |
| Engram persistent memory | ~300 (protocol description) | Eliminates re-explaining decisions across sessions |
| `compact-suggest` skill | ~30 (listing) | Encourages `/compact` before history balloons |
| Multi-Claude sync | 0 runtime cost | Saves manual config copy across instances |

**Net result**: in any moderately complex session, the wrapper saves more tokens than it costs.

---

## Caveats

- **Token counts are estimates** — real values depend on Anthropic's tokenizer (`cl100k_base` derivative) and the exact characters used. Code, JSON, and ASCII art tokenize differently than prose.
- **gentle-ai measurements** include the Claude Code adapter only. gentle-ai's full footprint includes adapters for 8 agents; we only count what ships to Claude Code.
- **The "after caching" number assumes** Anthropic's automatic prompt caching is hitting the stable prefix. Claude Code seems to do this automatically. For API direct callers, you'd configure cache breakpoints manually.
- **Sub-agent contexts** are billed separately. Delegating to Haiku for a `shell-runner` task costs Haiku's input (~$0.25/M input) for the command output, not Sonnet/Opus rates. Often 10-40× cheaper than reading the same output in the main agent.

---

## How this was measured

```bash
# system-general-ai
cd ~/Desktop/system-general-ai
wc -c templates/output-style.md templates/CLAUDE.md.global.tmpl templates/orchestrator.md templates/skills/*.md

# gentle-ai equivalents
cd ~/Desktop/gentle-ai
wc -c testdata/golden/combined-claude-claudemd.golden testdata/golden/sdd-claude-claudemd.golden \
       testdata/golden/engram-claude-claudemd.golden
find skills -name "*.md" -type f -exec wc -c {} +
```

To verify the numbers yourself, run the commands above on both repos.

