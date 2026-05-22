package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func testTemplates() fstest.MapFS {
	return fstest.MapFS{
		"codex/AGENTS.md.global.tmpl":    {Data: []byte("global rules")},
		"codex/AGENTS.md.project.tmpl":   {Data: []byte("project rules")},
		"codex/orchestrator.md":          {Data: []byte("sdd orchestrator")},
		"codex/engram-instructions.md":   {Data: []byte("engram instructions")},
		"codex/engram-compact-prompt.md": {Data: []byte("compact prompt")},
		"codex/skills/_shared/SKILL.md":  {Data: []byte("---\nname: _shared\ndescription: shared\n---\n")},
		"skills/sdd-init/SKILL.md":       {Data: []byte("---\nname: sdd-init\ndescription: init\n---\n")},
		"skills/shell-runner/SKILL.md":   {Data: []byte("---\nname: shell-runner\ndescription: shell\n---\n")},
		"output-style.md":                {Data: []byte("output style")},
	}
}

func TestSyncWritesCodexArtifactsAndPreservesUserConfig(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	inst := Instance{Path: filepath.Join(root, ".codex")}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(inst.Path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inst.ConfigPath(), []byte("model = \"gpt-5.5\"\n\n[projects.'x']\ntrust_level = \"trusted\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inst.AGENTSPath(), []byte("# User global\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Sync(inst, SyncOptions{
		Templates:    testTemplates(),
		ProjectDir:   project,
		EnableSDD:    true,
		EngramBinary: `C:\bin\engram.exe`,
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg := mustRead(t, inst.ConfigPath())
	for _, want := range []string{
		`model = "gpt-5.5"`,
		`[projects.'x']`,
		`model_instructions_file = "`,
		`experimental_compact_prompt_file = "`,
		`[mcp_servers.engram]`,
		`command = "C:\\bin\\engram.exe"`,
		`args = ["mcp", "--tools=agent"]`,
	} {
		if !strings.Contains(cfg, want) {
			t.Fatalf("config missing %q:\n%s", want, cfg)
		}
	}

	global := mustRead(t, inst.AGENTSPath())
	for _, want := range []string{"# User global", "global rules", "output style", "sdd orchestrator"} {
		if !strings.Contains(global, want) {
			t.Fatalf("global AGENTS missing %q:\n%s", want, global)
		}
	}

	projectAgents := mustRead(t, filepath.Join(project, "AGENTS.md"))
	if !strings.Contains(projectAgents, "project rules") {
		t.Fatalf("project AGENTS missing project rules:\n%s", projectAgents)
	}

	if _, err := os.Stat(filepath.Join(inst.SkillsDir(), "sdd-init", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(inst.SkillsDir(), "shell-runner", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(inst.EngramInstructionsPath()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(inst.Path, "backups")); err != nil {
		t.Fatal(err)
	}
}

func TestSyncIsIdempotentForManagedSections(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	inst := Instance{Path: filepath.Join(root, ".codex")}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	opts := SyncOptions{
		Templates:    testTemplates(),
		ProjectDir:   project,
		EnableSDD:    true,
		EngramBinary: "engram",
	}
	if err := Sync(inst, opts); err != nil {
		t.Fatal(err)
	}
	backupsAfterFirst := countBackupFiles(t, filepath.Join(inst.Path, "backups"))
	if err := Sync(inst, opts); err != nil {
		t.Fatal(err)
	}
	backupsAfterSecond := countBackupFiles(t, filepath.Join(inst.Path, "backups"))
	if backupsAfterSecond != backupsAfterFirst {
		t.Fatalf("second sync created backup without content change: before=%d after=%d", backupsAfterFirst, backupsAfterSecond)
	}

	global := mustRead(t, inst.AGENTSPath())
	if got := strings.Count(global, "<!-- BEGIN: system-general-ai/sdd-orchestrator -->"); got != 1 {
		t.Fatalf("sdd section count = %d; content:\n%s", got, global)
	}
	cfg := mustRead(t, inst.ConfigPath())
	if got := strings.Count(cfg, "[mcp_servers.engram]"); got != 1 {
		t.Fatalf("engram block count = %d; content:\n%s", got, cfg)
	}
}

func TestSyncConfigOnlyMutatesManagedTopLevelKeys(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	inst := Instance{Path: filepath.Join(root, ".codex")}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(inst.Path, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `# user comment
model = "gpt-5.5"
experimental_compact_prompt_file = "old-global.md"

[projects.'C:\repo\one']
model_instructions_file = "project-local.md"
trust_level = "trusted"

[mcp_servers.docs]
command = "docs-server"
args = ["serve"]
`
	if err := os.WriteFile(inst.ConfigPath(), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Sync(inst, SyncOptions{
		Templates:    testTemplates(),
		ProjectDir:   project,
		EnableSDD:    true,
		EngramBinary: `C:\bin\engram.exe`,
	}); err != nil {
		t.Fatal(err)
	}

	cfg := mustRead(t, inst.ConfigPath())
	for _, want := range []string{
		`# user comment`,
		`model = "gpt-5.5"`,
		`[projects.'C:\repo\one']`,
		`model_instructions_file = "project-local.md"`,
		`[mcp_servers.docs]`,
		`command = "docs-server"`,
		`[mcp_servers.engram]`,
		`command = "C:\\bin\\engram.exe"`,
	} {
		if !strings.Contains(cfg, want) {
			t.Fatalf("config missing preserved/generated entry %q:\n%s", want, cfg)
		}
	}
	if got := strings.Count(cfg, `experimental_compact_prompt_file = `); got != 1 {
		t.Fatalf("managed top-level compact prompt count = %d:\n%s", got, cfg)
	}
}

func TestSyncRejectsInvalidExistingConfig(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	inst := Instance{Path: filepath.Join(root, ".codex")}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(inst.Path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inst.ConfigPath(), []byte("this is not valid = [toml"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Sync(inst, SyncOptions{
		Templates:    testTemplates(),
		ProjectDir:   project,
		EnableSDD:    true,
		EngramBinary: "engram",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid TOML") {
		t.Fatalf("expected invalid TOML error, got %v", err)
	}
	if got := mustRead(t, inst.ConfigPath()); got != "this is not valid = [toml" {
		t.Fatalf("invalid config was modified:\n%s", got)
	}
	if _, statErr := os.Stat(inst.AGENTSPath()); !os.IsNotExist(statErr) {
		t.Fatalf("global AGENTS should not be written after invalid config, stat err=%v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(project, "AGENTS.md")); !os.IsNotExist(statErr) {
		t.Fatalf("project AGENTS should not be written after invalid config, stat err=%v", statErr)
	}
}

func TestUninstallRemovesOnlyManagedCodexArtifacts(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	inst := Instance{Path: filepath.Join(root, ".codex")}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Sync(inst, SyncOptions{
		Templates:    testTemplates(),
		ProjectDir:   project,
		EnableSDD:    true,
		EngramBinary: "engram",
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inst.SkillsDir(), "sdd-init", "notes.md"), []byte("user note"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Uninstall(inst, project, false); err != nil {
		t.Fatal(err)
	}

	if cfg, err := os.ReadFile(inst.ConfigPath()); err == nil {
		text := string(cfg)
		if strings.Contains(text, "mcp_servers.engram") || strings.Contains(text, "model_instructions_file") {
			t.Fatalf("managed config remained:\n%s", text)
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(inst.SkillsDir(), "sdd-init", "notes.md")); err != nil {
		t.Fatalf("user skill note should survive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(inst.SkillsDir(), "sdd-init", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("managed SKILL.md should be removed, err=%v", err)
	}
}

func countBackupFiles(t *testing.T, root string) int {
	t.Helper()
	count := 0
	err := filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return count
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
