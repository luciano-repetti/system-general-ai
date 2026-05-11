# Testing system-general-ai

Three layers of validation, in order of effort.

## Prerequisites

```bash
go version  # must be 1.24+
go mod tidy # downloads Cobra and indirect deps; run once after cloning
```

## Layer 1 — Unit tests (fast, automated)

```bash
go test ./internal/...
```

Coverage:

| Test | What it validates |
|---|---|
| `TestDetectInstances_FindsCLAUDEConfigDirEnv` | Env var detection (issue #253) |
| `TestDetectInstances_FindsClaudeWorkGlob` | Multi-Claude autodiscovery |
| `TestDetectInstances_DedupesEqualPaths` | No duplicates when env points to default |
| `TestApplyPersona_RemovesKeyOnNeutral` | **Fix #204 regression test** — neutral DELETES the outputStyle key |
| `TestApplyPersona_SetsKeyForNamed` | Normal persona switch |
| `TestApplyPersona_HandlesMissingFile` | Creates settings.json if absent |
| `TestSync_WritesAllExpectedFiles` | output-styles, CLAUDE.md, settings.json, 3+ skills |
| `TestSync_IsIdempotent` | Re-running Sync produces identical files |
| `TestSync_DisablingSDDRemovesOrchestratorSection` | Marker-based removal works |
| `TestSync_PreservesUserContentInCLAUDEMD` | User-authored content not clobbered |
| `TestSync_MergesSettingsPreservingUserKeys` | Shallow JSON merge respects user keys |
| `TestSelectAsset_PicksByOSArch` | Engram release asset selection |
| `TestSelectAsset_FailsWhenNoMatch` | Clean error on unsupported platform |

If any of these fail, **STOP** — there's a regression.

## Layer 2 — E2E smoke test (manual, isolated)

Builds the binary and runs `install` against an isolated tempdir. Does NOT touch your real `~/.claude`.

### Linux / macOS

```bash
chmod +x scripts/test-smoke.sh
./scripts/test-smoke.sh
```

The script verifies:
- Binary builds cleanly
- `install` produces all expected files
- CLAUDE.md has both marker sections
- settings.json has outputStyle, permissions, and `mcpServers.engram`
- All 11 skills are installed
- **Idempotency** — running install twice yields identical files
- `configure permissions strict` actually moves Edit/Write to ask
- `configure persona neutral` removes the outputStyle key (fix #204)
- `uninstall` removes managed files

### Windows

The smoke test is bash-only currently. To run on Windows:
- Use WSL, **or**
- Translate it to PowerShell (TODO — see `scripts/test-smoke.sh` for the logic)

## Layer 3 — Cross-platform build verification

Compiles for all 6 supported OS/arch combos. Doesn't run, just confirms it builds.

```bash
chmod +x scripts/test-build.sh
./scripts/test-build.sh
```

Output goes to `/tmp/sgai-builds/`. Ensures we ship binaries that link correctly for:
- linux/amd64, linux/arm64
- darwin/amd64 (Intel Mac), darwin/arm64 (Apple Silicon)
- windows/amd64, windows/arm64

## Manual end-to-end on your real Claude Code (optional)

After all the above pass, the truly final test is using it against a real Claude Code:

1. Build binary: `go build -o sgai ./cmd/system-general-ai`
2. **Backup** your current `~/.claude/settings.json` and `CLAUDE.md` first
3. Run: `./sgai install`
4. Open Claude Code, start a fresh session
5. Verify the Output Style is `system-general-ai` (the persona should match `templates/output-style.md`)
6. Verify CLAUDE.md has the marker sections
7. Verify `engram` is loaded (it should appear in MCP server list)
8. To restore: `./sgai uninstall` (and put back your backups if you had custom CLAUDE.md content)

## Troubleshooting

| Symptom | Fix |
|---|---|
| `cannot find package "github.com/spf13/cobra"` | Run `go mod tidy` first |
| `go.sum: missing go.sum entry` | Same — `go mod tidy` |
| Smoke test fails on `python3` | Replace with `python` or run JSON checks manually |
| Smoke test fails on idempotency | Check that `mergeMarkedSection` is using stable terminator characters |
| Tests pass but real install fails | Real install hits GitHub for engram release — needs internet |

