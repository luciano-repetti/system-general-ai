package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectInstances_FindsCLAUDEConfigDirEnv(t *testing.T) {
	tmp := t.TempDir()
	custom := filepath.Join(tmp, "custom-claude")
	if err := os.MkdirAll(custom, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", custom)
	t.Setenv("HOME", tmp)        // POSIX
	t.Setenv("USERPROFILE", tmp) // Windows

	got, err := DetectInstances()
	if err != nil {
		t.Fatalf("DetectInstances: %v", err)
	}

	found := false
	for _, inst := range got {
		if inst.Source == SourceEnv && inst.Path == custom {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find env-sourced instance at %s, got %v", custom, got)
	}
}

// TestDetectInstances_CLAUDEConfigDirIsExclusive guards against the regression
// that motivated this rule: when CLAUDE_CONFIG_DIR points at a single directory,
// autodiscovery (default ~/.claude and ~/.claude-work*) MUST be skipped.
// Otherwise a developer aiming an isolated test at a tempdir accidentally writes
// to every real Claude Code installation on the machine.
func TestDetectInstances_CLAUDEConfigDirIsExclusive(t *testing.T) {
	tmp := t.TempDir()

	// Create real-looking instances that should be IGNORED when env is set.
	for _, name := range []string{".claude", ".claude-work", ".claude-work-personal"} {
		if err := os.MkdirAll(filepath.Join(tmp, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// And the env target.
	custom := filepath.Join(tmp, "isolated-test")
	if err := os.MkdirAll(custom, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", custom)
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	got, err := DetectInstances()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("CLAUDE_CONFIG_DIR must be exclusive — expected 1 instance, got %d:\n%v", len(got), got)
	}
	if got[0].Path != custom {
		t.Fatalf("expected only %s, got %s", custom, got[0].Path)
	}
	if got[0].Source != SourceEnv {
		t.Fatalf("expected SourceEnv, got %s", got[0].Source)
	}
}

func TestDetectInstances_FindsClaudeWorkGlob(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("CLAUDE_CONFIG_DIR", "")

	for _, name := range []string{".claude-work", ".claude-work-personal", ".claude-work-laburo"} {
		if err := os.MkdirAll(filepath.Join(tmp, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := DetectInstances()
	if err != nil {
		t.Fatalf("DetectInstances: %v", err)
	}

	discovered := 0
	for _, inst := range got {
		if inst.Source == SourceDiscovered {
			discovered++
		}
	}
	if discovered != 3 {
		t.Fatalf("expected 3 discovered instances, got %d (%v)", discovered, got)
	}
}

func TestDetectInstances_DedupesEqualPaths(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(def, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("CLAUDE_CONFIG_DIR", def) // points at the same path as default

	got, err := DetectInstances()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, inst := range got {
		seen[inst.Path]++
	}
	if seen[def] != 1 {
		t.Fatalf("expected default path deduplicated, saw %d occurrences (%v)", seen[def], got)
	}
}
