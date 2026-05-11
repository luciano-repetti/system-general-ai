package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Source describes how an instance was found.
type Source string

const (
	SourceEnv        Source = "env"        // CLAUDE_CONFIG_DIR
	SourceDefault    Source = "default"    // $HOME/.claude
	SourceDiscovered Source = "discovered" // $HOME/.claude-work*
)

// Instance is one Claude Code installation we sync into.
type Instance struct {
	Path      string // absolute path to the instance dir
	IsDefault bool   // true for $HOME/.claude
	Source    Source
}

// SettingsPath returns the absolute path to settings.json for this instance.
func (i Instance) SettingsPath() string {
	return filepath.Join(i.Path, "settings.json")
}

// CLAUDEMDPath returns the absolute path to CLAUDE.md (global) for this instance.
func (i Instance) CLAUDEMDPath() string {
	return filepath.Join(i.Path, "CLAUDE.md")
}

// SkillsDir returns the absolute path to the skills directory for this instance.
func (i Instance) SkillsDir() string {
	return filepath.Join(i.Path, "skills")
}

// OutputStylesDir returns the absolute path to the output-styles directory.
func (i Instance) OutputStylesDir() string {
	return filepath.Join(i.Path, "output-styles")
}

// DetectInstances enumerates Claude Code installations on the system.
//
// Order of precedence:
//
//  1. CLAUDE_CONFIG_DIR — if set, this is treated as an EXCLUSIVE override:
//     only that directory is returned. The autodiscovery below is skipped.
//     This matches user expectation: pointing the variable at a single instance
//     should not also touch other detected installations.
//
//  2. Otherwise (env var unset or empty), the default + discovery paths are used:
//       $HOME/.claude            (default)
//       $HOME/.claude-work*      (multi-credentials, glob-discovered)
//
// Duplicates by absolute path are de-duplicated. Result is stable-sorted by Path.
func DetectInstances() ([]Instance, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home dir: %w", err)
	}

	seen := map[string]bool{}
	out := []Instance{}

	add := func(path string, isDefault bool, src Source) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return
		}
		if seen[abs] {
			return
		}
		if !isDir(abs) {
			return
		}
		seen[abs] = true
		out = append(out, Instance{Path: abs, IsDefault: isDefault, Source: src})
	}

	// 1. CLAUDE_CONFIG_DIR — EXCLUSIVE override when set.
	if env := os.Getenv("CLAUDE_CONFIG_DIR"); env != "" {
		add(env, false, SourceEnv)
		// Do NOT fall through to discovery: env override is exclusive.
		return out, nil
	}

	// 2. $HOME/.claude
	add(filepath.Join(home, ".claude"), true, SourceDefault)

	// 3. $HOME/.claude-work*
	matches, _ := filepath.Glob(filepath.Join(home, ".claude-work*"))
	for _, m := range matches {
		add(m, false, SourceDiscovered)
	}

	// Stable order
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})

	return out, nil
}

// FilterDefaults returns the instance Source name for a single CLAUDE_CONFIG_DIR override use.
// Helper kept for callers that want to act on a single instance only.
func FilterDefaults(instances []Instance) []Instance {
	out := []Instance{}
	for _, i := range instances {
		if i.IsDefault {
			out = append(out, i)
		}
	}
	return out
}

func isDir(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

// SourceLabel returns a human-friendly description for an Instance.
// Useful for installer reporting.
func SourceLabel(i Instance) string {
	parts := []string{string(i.Source)}
	if i.IsDefault {
		parts = append(parts, "default")
	}
	return strings.Join(parts, ",")
}
