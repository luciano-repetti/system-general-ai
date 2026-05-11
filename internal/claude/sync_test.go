package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// fakeTemplates returns an in-memory fs.FS shaped like our real templates folder.
func fakeTemplates() fstest.MapFS {
	return fstest.MapFS{
		"output-style.md":             {Data: []byte("# Output Style\nfake\n")},
		"CLAUDE.md.global.tmpl":       {Data: []byte("# Global Rules\nfake-rules\n")},
		"orchestrator.md":             {Data: []byte("# Orchestrator\nfake-orch\n")},
		"settings/permissive.json":    {Data: []byte(`{"outputStyle":"system-general-ai","permissions":{"allow":["Read"]}}`)},
		"settings/balanced.json":      {Data: []byte(`{"outputStyle":"system-general-ai","permissions":{"allow":["Read"],"ask":["Edit"]}}`)},
		"settings/strict.json":        {Data: []byte(`{"outputStyle":"system-general-ai","permissions":{"allow":["Read"],"ask":["Edit","Write"]}}`)},
		"skills/sdd-init/SKILL.md":        {Data: []byte("---\nname: sdd-init\n---\nfake init\n")},
		"skills/shell-runner/SKILL.md":    {Data: []byte("---\nname: shell-runner\n---\nfake runner\n")},
		"skills/compact-suggest/SKILL.md": {Data: []byte("---\nname: compact-suggest\n---\nfake compact\n")},
	}
}

func TestSync_WritesAllExpectedFiles(t *testing.T) {
	tmp := t.TempDir()
	inst := Instance{Path: tmp, IsDefault: true, Source: SourceDefault}

	err := Sync(inst, SyncOptions{
		Templates:    fakeTemplates(),
		Mode:         ModePermissive,
		EnableSDD:    true,
		EngramBinary: "/tmp/engram",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	expected := []string{
		"output-styles/system-general-ai.md",
		"CLAUDE.md",
		"settings.json",
		"skills/sdd-init/SKILL.md",
		"skills/shell-runner/SKILL.md",
		"skills/compact-suggest/SKILL.md",
	}
	for _, rel := range expected {
		if _, err := os.Stat(filepath.Join(tmp, rel)); err != nil {
			t.Errorf("expected %s, missing: %v", rel, err)
		}
	}

	claudeMD, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(claudeMD), "<!-- BEGIN: system-general-ai/global-rules -->") {
		t.Error("global-rules marker missing in CLAUDE.md")
	}
	if !strings.Contains(string(claudeMD), "<!-- BEGIN: system-general-ai/sdd-orchestrator -->") {
		t.Error("sdd-orchestrator marker missing in CLAUDE.md")
	}
}

func TestSync_IsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	inst := Instance{Path: tmp, IsDefault: true, Source: SourceDefault}
	opts := SyncOptions{
		Templates: fakeTemplates(),
		Mode:      ModePermissive,
		EnableSDD: true,
	}

	if err := Sync(inst, opts); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	first, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	if err := Sync(inst, opts); err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	second, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	if string(first) != string(second) {
		t.Errorf("Sync not idempotent — content changed between runs.\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestSync_DisablingSDDRemovesOrchestratorSection(t *testing.T) {
	tmp := t.TempDir()
	inst := Instance{Path: tmp, IsDefault: true, Source: SourceDefault}
	opts := SyncOptions{
		Templates: fakeTemplates(),
		Mode:      ModePermissive,
		EnableSDD: true,
	}

	if err := Sync(inst, opts); err != nil {
		t.Fatal(err)
	}

	opts.EnableSDD = false
	if err := Sync(inst, opts); err != nil {
		t.Fatal(err)
	}

	claudeMD, _ := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if strings.Contains(string(claudeMD), "<!-- BEGIN: system-general-ai/sdd-orchestrator -->") {
		t.Error("orchestrator section was NOT removed when EnableSDD=false")
	}
	if !strings.Contains(string(claudeMD), "<!-- BEGIN: system-general-ai/global-rules -->") {
		t.Error("global-rules section was lost when removing orchestrator")
	}
}

func TestSync_PreservesUserContentInCLAUDEMD(t *testing.T) {
	tmp := t.TempDir()
	inst := Instance{Path: tmp, IsDefault: true, Source: SourceDefault}

	userContent := "# My Project\n\nUser-authored stuff that must not be deleted.\n"
	if err := os.WriteFile(filepath.Join(tmp, "CLAUDE.md"), []byte(userContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Sync(inst, SyncOptions{
		Templates: fakeTemplates(),
		Mode:      ModePermissive,
		EnableSDD: true,
	}); err != nil {
		t.Fatal(err)
	}

	claudeMD, _ := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if !strings.Contains(string(claudeMD), "User-authored stuff that must not be deleted.") {
		t.Error("user-authored content was clobbered")
	}
}

func TestSync_MergesSettingsPreservingUserKeys(t *testing.T) {
	tmp := t.TempDir()
	inst := Instance{Path: tmp, IsDefault: true, Source: SourceDefault}

	userSettings := `{"theme":"dark","editor":{"fontSize":14}}`
	if err := os.WriteFile(filepath.Join(tmp, "settings.json"), []byte(userSettings), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Sync(inst, SyncOptions{
		Templates: fakeTemplates(),
		Mode:      ModePermissive,
		EnableSDD: true,
	}); err != nil {
		t.Fatal(err)
	}

	got := readJSONFile(t, filepath.Join(tmp, "settings.json"))
	if got["theme"] != "dark" {
		t.Errorf("user 'theme' lost: %v", got["theme"])
	}
	if got["outputStyle"] != "system-general-ai" {
		t.Errorf("our outputStyle not applied: %v", got["outputStyle"])
	}
}

