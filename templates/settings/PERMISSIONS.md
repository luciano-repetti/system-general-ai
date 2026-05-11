# Permission modes — system-general-ai

The installer ships three permission profiles. Default is **permissive**.

## How to switch modes

```bash
system-general-ai configure permissions <strict|balanced|permissive>
```

Or pass the flag during install:

```bash
system-general-ai install --strictness=<strict|balanced|permissive>
```

This rewrites the relevant `permissions` block in `~/.claude/settings.json` for every detected Claude Code instance. Manual edits to the file are preserved when they don't conflict with the chosen mode (the installer merges; it doesn't overwrite arbitrary keys).

---

## permissive (default)

**Productivity-first. Recommended.**

- Allows: Read, Edit, Write, Grep, Glob, Bash (with safety denylist), WebFetch, WebSearch, Task system, Agent, Skill, MCP tools
- Denies: actually destructive operations (rm -rf on root or home, force-push, hard reset, fork bombs, dd, mkfs, branch deletion of main/master)

The agent works without interruptions for safe operations. Only truly destructive commands are blocked.

## balanced

**Confirmation for state-changing operations.**

- Allows: Read, Grep, Glob, Task system, Skill, MCP memory tools, common safe Bash commands (git status/diff/log, ls, cat, head, tail, etc)
- Asks: Edit, Write, arbitrary Bash, WebFetch, WebSearch, Agent, NotebookEdit
- Denies: same destructive list as permissive

Best when you want to review every file edit before it lands.

## strict

**Read-only-ish by default.**

- Allows: Read, Grep, Glob, TaskGet, TaskList, ToolSearch, minimal Bash (git status/diff/log, pwd, ls), MCP memory search/get
- Asks: everything else, including Skill, MCP write tools
- Denies: same destructive list as permissive

Best for sensitive codebases (production secrets, regulated environments) or when teaching the agent a new project and you want full visibility.

---

## Customizing further

The shipped JSON files are starting points. Edit `~/.claude/settings.json` directly to add your own allow/ask/deny patterns. The installer never overwrites your manual additions — it only manages the keys it owns.

Pattern syntax follows Claude Code conventions:
- `Tool` → matches the tool name
- `Tool(arg)` → matches a specific argument
- `Tool(arg*)` → matches any argument starting with `arg`
- `Tool(*pattern*)` → matches any argument containing `pattern`

## Sync across multi-Claude instances

When you change the mode (or add custom rules), the installer propagates the change to every detected `.claude-work*` instance under `$HOME`. No manual copying.

