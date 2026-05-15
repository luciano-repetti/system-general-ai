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
	if err := Sync(inst, opts); err != nil {
		t.Fatal(err)
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

func mustRead(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
