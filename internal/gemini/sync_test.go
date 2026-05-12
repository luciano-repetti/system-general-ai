package gemini

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func fakeTemplates() fstest.MapFS {
	return fstest.MapFS{
		"templates/gemini/GEMINI.md.global.tmpl": {Data: []byte("# Global Rules\nfake-rules\n")},
		"templates/gemini/orchestrator.md":       {Data: []byte("# Orchestrator\nfake-orch\n")},
		"templates/gemini/skills/sdd-init/SKILL.md":  {Data: []byte("---\nname: sdd-init\n---\nfake init\n")},
	}
}

func TestSync_ProjectRules(t *testing.T) {
	tmp := t.TempDir()
	opts := SyncOptions{
		Templates: fakeTemplates(),
		EnableSDD: true,
	}

	if err := syncProjectRules(tmp, opts); err != nil {
		t.Fatalf("syncProjectRules: %v", err)
	}

	geminiMD, err := os.ReadFile(filepath.Join(tmp, "GEMINI.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(geminiMD), "<!-- BEGIN: system-general-ai/global-rules -->") {
		t.Error("global-rules marker missing")
	}
	if !strings.Contains(string(geminiMD), "<!-- BEGIN: system-general-ai/sdd-orchestrator -->") {
		t.Error("sdd-orchestrator marker missing")
	}
}

func TestSync_Idempotency(t *testing.T) {
	tmp := t.TempDir()
	opts := SyncOptions{
		Templates: fakeTemplates(),
		EnableSDD: true,
	}

	if err := syncProjectRules(tmp, opts); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(filepath.Join(tmp, "GEMINI.md"))

	if err := syncProjectRules(tmp, opts); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(filepath.Join(tmp, "GEMINI.md"))

	if string(first) != string(second) {
		t.Error("syncProjectRules not idempotent")
	}
}

func TestRemoveMarkedSection(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "GEMINI.md")
	content := "User stuff\n<!-- BEGIN: system-general-ai/test -->\nDelete me\n<!-- END: system-general-ai/test -->\nMore user stuff\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := RemoveMarkedSection(path, "test"); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(path)
	if strings.Contains(string(got), "Delete me") {
		t.Error("section not removed")
	}
	if !strings.Contains(string(got), "User stuff") || !strings.Contains(string(got), "More user stuff") {
		t.Error("user content lost")
	}
}
