# Troubleshooting

Real failure modes with concrete fixes. Source references point at the code that produces the behavior.

## Multi-Claude not detected

Symptom: `sgai install` reports `no Claude Code installation found at ~/.claude or any .claude-work*. Install Claude Code first` and exits.

Detection rules (`internal/claude/detect.go:61`):
- If `CLAUDE_CONFIG_DIR` is set and non-empty, ONLY that path is considered (exclusive override). Even if it doesn't exist as a directory, autodiscovery is skipped.
- Otherwise: `~/.claude` is the default; `~/.claude-work*` is autodiscovered.

Fixes:
1. Run `echo $CLAUDE_CONFIG_DIR` (POSIX) or `$env:CLAUDE_CONFIG_DIR` (PowerShell). If set, verify the path exists with `eza` / `Get-ChildItem`. Unset it (`unset CLAUDE_CONFIG_DIR` / `Remove-Item Env:CLAUDE_CONFIG_DIR`) to fall back to default + autodiscovery.
2. Verify `~/.claude` is a real directory, not a missing path or a regular file. `isDir` (`internal/claude/detect.go:121`) requires `os.Stat` to return a directory.
3. If you intentionally manage instances under non-standard paths, set `CLAUDE_CONFIG_DIR` to the one you want for this run.

## Persona didn't switch

Symptom: ran `sgai configure persona <name>` (or `-` for neutral), but Claude Code still uses the old persona.

Fixes:
1. Restart Claude Code. `settings.json` is read at startup.
2. Verify `<instance>/settings.json` has the expected `outputStyle` value. Use `bat` to inspect.
3. **Neutral case (bug fix #204)**: when you switched to neutral via `-` or `neutral`, the `outputStyle` key must be **absent** from `settings.json`, not set to `""`. `ApplyPersona` (`internal/claude/persona.go:32`) deletes the key. If you see `"outputStyle": ""`, that's a stale-key bug from another tool — re-run `sgai configure persona neutral` to delete it.

## Permissions look weird after install

Symptom: after `sgai install` over an existing instance, your custom `permissions.allow` entries vanished or `defaultMode` flipped.

Pre-fix behavior: `mergeJSONFile` did a shallow top-level replacement, so the entire `permissions` object got overwritten. Fixed by deep-merging via `mergePermissions` (`internal/claude/sync.go:309`).

Contract now:
- `allow`/`deny`/`ask` are deduplicated UNION (template-first for stable order)
- `defaultMode` — user value wins if set, otherwise template value
- Other sub-keys — user value wins

Fixes:
1. Upgrade to a binary built after the post-launch fix. `sgai --version` and check against the release notes.
2. Verify with `bat <instance>/settings.json`: count entries in `permissions.allow` before vs after `sgai sync`. Custom entries should survive.
3. If you're stuck on an old binary and lost custom entries, restore them manually and upgrade before running `sync` again.

## Skill name collisions

sgai owns these 11 skill names under `<instance>/skills/`:

```
sdd-init   sdd-explore   sdd-propose   sdd-spec   sdd-design
sdd-tasks  sdd-apply     sdd-verify    sdd-archive
shell-runner   compact-suggest
```

If you maintain your own skill at one of these names, `sgai install` / `sgai sync` overwrites the `SKILL.md` inside that directory. By design.

`syncSkills` (`internal/claude/sync.go:150`) writes only `SKILL.md`. Other files inside the same skill directory (notes, images, supporting scripts) are preserved. `uninstall` removes only `SKILL.md` and then removes the directory only if empty (`internal/cli/uninstall.go:91`).

Workaround: rename your custom skill to a non-colliding name.

## Multi-profile hardlinks (Windows or POSIX)

Symptom: you maintain `~/.claude` and `~/.claude-work2` as separate profiles, but their `CLAUDE.md` (or other config files) are hardlinked across profiles. Running `sgai install` writes through the hardlink and changes appear in BOTH profiles.

Cause: a hardlink points multiple paths at the same on-disk inode. Any tool that opens the file by path and writes to it modifies the shared content.

Fix: break the hardlink before installing. The procedure:

POSIX (`bash`):
```bash
cp -p ~/.claude-work2/CLAUDE.md ~/.claude-work2/CLAUDE.md.tmp
rm ~/.claude-work2/CLAUDE.md
mv ~/.claude-work2/CLAUDE.md.tmp ~/.claude-work2/CLAUDE.md
```

Windows (PowerShell):
```powershell
$src = "$env:USERPROFILE\.claude-work2\CLAUDE.md"
Copy-Item $src "$src.tmp"
Remove-Item $src
Move-Item "$src.tmp" $src
```

Apply the same procedure to `settings.json`, `output-styles/system-general-ai.md`, and any skill `SKILL.md` that's hardlinked. Then run `sgai sync`.

## Junctioned subdirectories

Symptom (Windows): `~/.claude-work2/skills/`, `~/.claude-work2/output-styles/`, `~/.claude-work2/mcp/`, or `~/.claude-work2/plugins/` is an NTFS junction pointing back at `~/.claude/`. Running `CLAUDE_CONFIG_DIR=~/.claude-work2 sgai install` (or letting autodiscovery handle it) writes files into the junction target — i.e. into `~/.claude/`. The two profiles aren't actually isolated.

Detect: `Get-Item <path> | Select Mode, LinkType, Target`. `LinkType: Junction` means it points elsewhere.

Same hardlink-style mental model applies: anything you write under a junction lands at the target. To isolate a profile, replace the junction with a real directory (copy the contents over, remove the junction with `Remove-Item -Recurse`, recreate as a real directory) before running `sgai install` or `sgai sync`.

## Build from source

Requires **Go 1.24+**.

```bash
git clone https://github.com/luciano-repetti/system-general-ai
cd system-general-ai
go mod tidy
go build -o sgai ./cmd/system-general-ai
./sgai install
```

Validate with `scripts/test-smoke.sh` — runs in an isolated tempdir and never touches your real `~/.claude`. Now Windows-portable (Bug 3 in `PLAN.md`).

## Smoke test fails with python errors on Windows

Should not happen on the current code. The script (`scripts/test-smoke.sh`) detects the Python interpreter via a `$PY` resolver and converts paths through `cygpath -m` (Bug 3 fix, 2026-05).

If it still fails:
- Ensure `python` or `python3` is on `PATH` (`which python` / `Get-Command python`).
- Ensure `cygpath` is available — Git Bash for Windows ships with it. Native PowerShell does not; run the script from Git Bash, not from `cmd.exe` or PowerShell directly.
- If `python` resolves to the Microsoft Store stub (`%LOCALAPPDATA%\Microsoft\WindowsApps\python.exe`), install a real Python distribution and put it ahead on `PATH`.

