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

func TestSync(t *testing.T) {
	projectDir := t.TempDir()
	home := t.TempDir()
	t.Setenv("USERPROFILE", home) // For Windows
	t.Setenv("HOME", home)        // For Unix

	opts := SyncOptions{
		Templates: fakeTemplates(),
		EnableSDD: true,
	}

	if err := Sync(projectDir, opts); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// 1. Check project GEMINI.md
	projectRulesPath := filepath.Join(projectDir, "GEMINI.md")
	projectMD, err := os.ReadFile(projectRulesPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(projectMD), "<!-- BEGIN: system-general-ai/global-rules -->") {
		t.Error("global-rules marker missing in project rules")
	}
	if !strings.Contains(string(projectMD), "<!-- BEGIN: system-general-ai/sdd-orchestrator -->") {
		t.Error("sdd-orchestrator marker missing in project rules")
	}

	// 2. Check global rules
	globalRulesPath := filepath.Join(home, ".gemini", "GEMINI.md")
	if _, err := os.Stat(globalRulesPath); os.IsNotExist(err) {
		t.Errorf("global GEMINI.md not created at %s", globalRulesPath)
	}

	// 3. Sync with persona
	opts.Persona = "Test Persona"
	if err := Sync(projectDir, opts); err != nil {
		t.Fatalf("Sync with persona failed: %v", err)
	}

	// Check persona in project rules
	content, _ := os.ReadFile(projectRulesPath)
	if !strings.Contains(string(content), "## Persona") || !strings.Contains(string(content), "Test Persona") {
		t.Errorf("Persona section missing or incorrect in project rules: %s", string(content))
	}

	// Check persona in global rules
	globalContent, _ := os.ReadFile(globalRulesPath)
	if !strings.Contains(string(globalContent), "## Persona") || !strings.Contains(string(globalContent), "Test Persona") {
		t.Errorf("Persona section missing or incorrect in global rules")
	}

	// 4. Sync with neutral persona (remove)
	opts.Persona = ""
	if err := Sync(projectDir, opts); err != nil {
		t.Fatalf("Sync with neutral persona failed: %v", err)
	}
	content, _ = os.ReadFile(projectRulesPath)
	if strings.Contains(string(content), "## Persona") {
		t.Errorf("Persona section should have been removed")
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
