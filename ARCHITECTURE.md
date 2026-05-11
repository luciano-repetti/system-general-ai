# Architecture Decisions — system-general-ai

> Decisions made during Stage 1. Each decision has rationale and trade-offs documented so future contributors (or future-us) can understand the reasoning.

---

## Language: Go 1.24+

**Decision**: Go is the primary language.

**Rationale**:
- **Fast development** — simpler than Rust for CLI/TUI work; no borrow checker overhead
- **Mature ecosystem** for CLI work: Cobra, Bubble Tea, Viper, Lipgloss all production-grade
- **Native cross-compilation**: `GOOS=windows GOARCH=amd64 go build` works out of the box
- **Single static binary distribution** — no runtime needed on the user's machine

**Rejected alternatives**:
- **Rust** — marginally better runtime, but development time is significantly longer. For a CLI/installer, runtime differences are imperceptible to the user. Rust valuable in high-concurrency services, not here.
- **Node/TypeScript** — would have made `pnpm i -g` natural, but requires runtime on every user's machine. Distribution complexity moves from us to the user.
- **Pure shell** — too limited for multi-Claude detection, MCP config generation, and template rendering.

**Min Go version**: 1.24.

---

## Distribution

**Decision**: Pre-compiled binaries via GitHub Releases + bootstrap install scripts. Package managers as Stage 9 bonus.

**Primary install paths**:

| Platform | Method |
|---|---|
| Linux/macOS | `curl -fsSL <url>/install.sh \| bash` |
| Windows | `irm <url>/install.ps1 \| iex` |
| Any with Go 1.24+ | `go install github.com/<org>/system-general-ai/cmd/system-general-ai@latest` |

**Stage 9 additions**:
- Scoop bucket for Windows
- Homebrew tap for macOS/Linux
- Optional: scoop manifest, brew formula, winget package

**Rejected**:
- **`pnpm i -g`** — incompatible with Go binary distribution.
- **Manual binary download + alias setup** — defeats the "single command" promise; users don't want to fiddle with PATH.

---

## CLI Framework

**Decision**: Cobra (`github.com/spf13/cobra`).

**Rationale**:
- De-facto standard for Go CLIs (kubectl, hugo, gh, terraform all use it)
- Auto-generated help, completion, man pages
- Composable subcommands match the pattern we need (`system-general-ai install`, `... sync`, `... doctor`, etc)
- Active maintenance

**Rejected**:
- **stdlib `flag`** — too low-level for nested subcommands
- **urfave/cli** — fine but smaller ecosystem and less convention
- **kingpin** — abandoned

---

## TUI: Optional, Bubble Tea if needed

**Decision**: NO mandatory TUI. Bubble Tea + Lipgloss available for optional interactive flows.

**Rationale**:
- Upstream issue #95 documents how a mandatory Bubble Tea TUI crashes in headless/no-TTY environments
- Smart defaults must work without ANY interaction: `system-general-ai install` with zero flags should produce a working installation
- Interactive TUI only used for: explicit configuration commands (`system-general-ai configure`), explicit selection of multi-Claude instances when ambiguous

**TTY detection**: Check `isatty` before launching any interactive UI; fall back to non-interactive mode otherwise.

---

## Configuration Format

**Decision**: YAML for user-facing config files. Generated `CLAUDE.md` and skill files stay as Markdown (Claude Code's expected format).

**Rationale**:
- YAML fits orchestration metadata naturally (`settings.json` is the exception — Claude Code's own format)
- Human-editable and diffable
- Standard in the Go ecosystem for config files

**Library**: `gopkg.in/yaml.v3` (stable, well-maintained).

---

## Engram Integration

**Decision**: Consume the existing Engram binary directly. NO vendor, NO fork.

**Rationale**:
- Engram is the persistent-memory MCP server we use; it's MIT-licensed and ships as a standalone Go binary via GitHub Releases
- Forking would mean maintaining a parallel codebase for no benefit — its functionality is exactly what we need
- We download the latest release at install time and configure Claude Code's MCP block to use it

**Risk**: If upstream Engram makes a breaking change, our installer detects the version mismatch and warns. Worst case, we pin to a known-good version.

**Install location**: `$HOME/.local/bin/engram` (Linux/macOS) or `%LOCALAPPDATA%\system-general-ai\bin\engram.exe` (Windows). PATH is updated by the installer.

---

## Project Structure

```
system-general-ai/
├── cmd/
│   └── system-general-ai/        # main entry point — main.go
├── internal/                     # private packages, not importable externally
│   ├── installer/                # install/uninstall logic
│   ├── claude/                   # Claude Code instance detection (multi-claude)
│   ├── engram/                   # engram binary download + MCP config generation
│   ├── sdd/                      # SDD orchestrator templates and skill generation
│   ├── templates/                # template rendering (CLAUDE.md, skills, etc)
│   ├── tui/                      # optional Bubble Tea views
│   └── system/                   # OS detection, PATH manipulation, file ops
├── pkg/                          # (reserved for public packages if any)
├── templates/                    # raw template files copied to .claude/ at install
│   ├── CLAUDE.md.tmpl            # global CLAUDE.md
│   ├── output-style.md           # custom system prompt (Stage 2)
│   ├── skills/                   # SDD phase skills
│   └── mcp/                      # MCP server config snippets
├── scripts/
│   ├── install.sh                # Linux/macOS bootstrap
│   └── install.ps1               # Windows bootstrap
├── docs/
│   └── ...                       # user-facing documentation
├── testdata/
│   └── golden/                   # golden file tests for template generation
├── go.mod
├── go.sum
├── README.md
├── PLAN.md
├── TASKS.md
├── ARCHITECTURE.md               # this file
└── LICENSE                       # MIT
```

**Conventions**:
- `internal/` — private, not importable from outside
- `cmd/` — executables
- `templates/` — files distributed verbatim at install time, NOT compiled in (allows hot-reload during dev)
- `testdata/golden/` — captured outputs for regression tests on template generation

---

## Testing

**Decision**: Go's stdlib `testing` package + golden file tests.

**Test categories**:
- **Unit tests** — pure functions, parsers, generators
- **Golden tests** — assert that generated CLAUDE.md / skill files match captured snapshots
- **Integration tests** — actual install in a temp directory, verify PATH/files are set up
- **E2E tests** — light, cover the install→sync→uninstall cycle on each OS in CI

**Frameworks rejected**:
- **Ginkgo/Gomega** — overkill for this scope
- **testify** — fine, but stdlib is enough

---

## Versioning & Release

**Decision**: Semantic versioning. Tagged releases on GitHub. Changelog auto-generated from conventional commits.

**Release artifacts** (per tag):
- Binary for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`
- SHA256 checksums file
- Source tarball (auto)

**Build automation**: GitHub Actions matrix on `goreleaser`.

---

## License & Credits

- **License**: MIT
- **Engram**: documented as upstream dependency with link to its repo

---

## Open architecture questions (deferred)

Items we'll decide when we get to the relevant stage:

- **Stage 4**: should SDD phase skills be generated dynamically (Go templates) or shipped as static markdown? Trade-off: dynamic = more flexible at the cost of complexity; static = simpler but harder to keep DRY
- **Stage 6**: how to detect `.claude-work*` instances reliably across OS path conventions
- **Stage 8**: RTK implementation — fork the existing tool or write a minimal version
- **Stage 9**: scoop manifest hosting — same repo or separate `scoop-bucket` repo

