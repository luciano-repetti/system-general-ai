package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestApplyPersona_RemovesKeyOnNeutral is a regression test:
// switching to neutral must DELETE the outputStyle key, not set it to "".
func TestApplyPersona_RemovesKeyOnNeutral(t *testing.T) {
	tmp := t.TempDir()
	settings := filepath.Join(tmp, "settings.json")

	initial := map[string]any{
		"outputStyle": "previous-style",
		"theme":       "dark",
	}
	writeJSONFile(t, settings, initial)

	if err := ApplyPersona(settings, PersonaNeutral); err != nil {
		t.Fatalf("ApplyPersona: %v", err)
	}

	got := readJSONFile(t, settings)
	if _, present := got["outputStyle"]; present {
		t.Fatalf("outputStyle key was NOT removed, got %v", got["outputStyle"])
	}
	if got["theme"] != "dark" {
		t.Fatalf("user-owned key 'theme' was clobbered, got %v", got["theme"])
	}
}

func TestApplyPersona_SetsKeyForNamed(t *testing.T) {
	tmp := t.TempDir()
	settings := filepath.Join(tmp, "settings.json")
	writeJSONFile(t, settings, map[string]any{"theme": "dark"})

	if err := ApplyPersona(settings, "system-general-ai"); err != nil {
		t.Fatalf("ApplyPersona: %v", err)
	}

	got := readJSONFile(t, settings)
	if got["outputStyle"] != "system-general-ai" {
		t.Fatalf("outputStyle not set, got %v", got["outputStyle"])
	}
	if got["theme"] != "dark" {
		t.Fatalf("user key clobbered: %v", got["theme"])
	}
}

func TestApplyPersona_HandlesMissingFile(t *testing.T) {
	tmp := t.TempDir()
	settings := filepath.Join(tmp, "settings.json") // does not exist

	if err := ApplyPersona(settings, "system-general-ai"); err != nil {
		t.Fatalf("ApplyPersona on missing file: %v", err)
	}

	got := readJSONFile(t, settings)
	if got["outputStyle"] != "system-general-ai" {
		t.Fatalf("outputStyle not set after creating file")
	}
}

// helpers

func writeJSONFile(t *testing.T, path string, data map[string]any) {
	t.Helper()
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
