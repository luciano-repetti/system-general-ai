package claude

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// PermissionMode is one of the three shipped permission profiles.
type PermissionMode string

const (
	ModePermissive PermissionMode = "permissive"
	ModeBalanced   PermissionMode = "balanced"
	ModeStrict     PermissionMode = "strict"
)

// SyncOptions controls what gets pushed to a single Claude Code instance.
type SyncOptions struct {
	// Templates is the filesystem (typically an embed.FS) rooted at the templates
	// folder that ships with system-general-ai. The expected layout is:
	//   output-style.md
	//   CLAUDE.md.global.tmpl
	//   orchestrator.md
	//   settings/{permissive,balanced,strict}.json
	//   skills/<skill>/SKILL.md
	Templates fs.FS

	// Mode is the permission profile to apply (permissive | balanced | strict).
	Mode PermissionMode

	// EnableSDD controls whether the orchestrator block is appended to CLAUDE.md.
	// Skills are always synced regardless of this flag (shell-runner, compact-suggest
	// must be available even when SDD is off).
	EnableSDD bool

	// EngramBinary is the absolute path to the engram binary (from internal/engram.Install).
	// If empty, the MCP block for engram is not written.
	EngramBinary string
}

// Sync applies the templates + settings + MCP config to a single instance.
// It is idempotent: re-running with the same options produces the same result.
// User-authored content is preserved. Sections we own are delimited with HTML markers.
func Sync(inst Instance, opts SyncOptions) error {
	if err := ensureDirs(inst); err != nil {
		return err
	}

	if err := syncOutputStyle(inst, opts); err != nil {
		return fmt.Errorf("output style: %w", err)
	}

	if err := syncCLAUDEMD(inst, opts); err != nil {
		return fmt.Errorf("CLAUDE.md: %w", err)
	}

	if err := syncSettings(inst, opts); err != nil {
		return fmt.Errorf("settings.json: %w", err)
	}

	// Skills are always synced — `shell-runner` must be available even when SDD is disabled.
	// SDD-specific skills (sdd-*) coexist harmlessly when SDD is off because the orchestrator
	// block is not in CLAUDE.md, so the agent never invokes them.
	if err := syncSkills(inst, opts); err != nil {
		return fmt.Errorf("skills: %w", err)
	}

	if opts.EngramBinary != "" {
		if err := syncEngramMCP(inst, opts.EngramBinary); err != nil {
			return fmt.Errorf("engram MCP: %w", err)
		}
	}

	return nil
}

func ensureDirs(inst Instance) error {
	dirs := []string{
		inst.Path,
		inst.SkillsDir(),
		inst.OutputStylesDir(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}
	return nil
}

func syncOutputStyle(inst Instance, opts SyncOptions) error {
	dst := filepath.Join(inst.OutputStylesDir(), "system-general-ai.md")
	return copyFromFS(opts.Templates, "output-style.md", dst)
}

func syncCLAUDEMD(inst Instance, opts SyncOptions) error {
	globalContent, err := fs.ReadFile(opts.Templates, "CLAUDE.md.global.tmpl")
	if err != nil {
		return fmt.Errorf("read global tmpl: %w", err)
	}
	if err := mergeMarkedSection(inst.CLAUDEMDPath(), "global-rules", string(globalContent)); err != nil {
		return err
	}

	if opts.EnableSDD {
		orchContent, err := fs.ReadFile(opts.Templates, "orchestrator.md")
		if err != nil {
			return fmt.Errorf("read orchestrator: %w", err)
		}
		if err := mergeMarkedSection(inst.CLAUDEMDPath(), "sdd-orchestrator", string(orchContent)); err != nil {
			return err
		}
	} else {
		if err := removeMarkedSection(inst.CLAUDEMDPath(), "sdd-orchestrator"); err != nil {
			return err
		}
	}

	return nil
}

func syncSettings(inst Instance, opts SyncOptions) error {
	srcPath := path.Join("settings", string(opts.Mode)+".json")
	raw, err := fs.ReadFile(opts.Templates, srcPath)
	if err != nil {
		return fmt.Errorf("read template %s: %w", srcPath, err)
	}

	var template map[string]any
	if err := json.Unmarshal(raw, &template); err != nil {
		return fmt.Errorf("parse template settings: %w", err)
	}

	owned := map[string]any{}
	for _, k := range []string{"outputStyle", "permissions", "$schema"} {
		if v, ok := template[k]; ok {
			owned[k] = v
		}
	}

	return mergeJSONFile(inst.SettingsPath(), owned)
}

func syncSkills(inst Instance, opts SyncOptions) error {
	// Skills now ship as directories: templates/skills/<name>/SKILL.md
	// Each gets installed to <SkillsDir>/<name>/SKILL.md (Claude Code's
	// directory-format skill convention).
	entries, err := fs.ReadDir(opts.Templates, "skills")
	if err != nil {
		return fmt.Errorf("read skills dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		srcPath := path.Join("skills", e.Name(), "SKILL.md")
		dst := filepath.Join(inst.SkillsDir(), e.Name(), "SKILL.md")
		if err := copyFromFS(opts.Templates, srcPath, dst); err != nil {
			return fmt.Errorf("copy skill %s: %w", e.Name(), err)
		}
	}
	return nil
}

func syncEngramMCP(inst Instance, engramBinary string) error {
	// Inline build of the engram MCP config (avoid importing engram package to keep claude/ self-contained).
	cfg := map[string]any{
		"command": engramBinary,
		"args":    []string{"serve"},
		"env":     map[string]string{},
	}
	return mergeMCPServer(inst.SettingsPath(), "engram", cfg)
}

// --- helpers ---

func copyFromFS(srcFS fs.FS, srcPath, dst string) error {
	in, err := srcFS.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

const (
	markerStartFmt = "<!-- BEGIN: system-general-ai/%s -->"
	markerEndFmt   = "<!-- END: system-general-ai/%s -->"
)

func mergeMarkedSection(path, sectionName, content string) error {
	startMarker := fmt.Sprintf(markerStartFmt, sectionName)
	endMarker := fmt.Sprintf(markerEndFmt, sectionName)
	newSection := fmt.Sprintf("%s\n%s\n%s", startMarker, strings.TrimSpace(content), endMarker)

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(existing) == 0 {
		return writeFileEnsureDir(path, []byte(newSection+"\n"))
	}

	existingStr := string(existing)
	startIdx := strings.Index(existingStr, startMarker)
	endIdx := strings.Index(existingStr, endMarker)

	if startIdx >= 0 && endIdx > startIdx {
		before := existingStr[:startIdx]
		after := existingStr[endIdx+len(endMarker):]
		merged := before + newSection + after
		return writeFileEnsureDir(path, []byte(merged))
	}

	sep := "\n\n"
	if strings.HasSuffix(existingStr, "\n\n") {
		sep = ""
	} else if strings.HasSuffix(existingStr, "\n") {
		sep = "\n"
	}
	merged := existingStr + sep + newSection + "\n"
	return writeFileEnsureDir(path, []byte(merged))
}

// RemoveMarkedSection deletes a marker-delimited section from a markdown file.
// Idempotent: returns nil if the section or the file doesn't exist.
func RemoveMarkedSection(path, sectionName string) error {
	return removeMarkedSection(path, sectionName)
}

func removeMarkedSection(path, sectionName string) error {
	startMarker := fmt.Sprintf(markerStartFmt, sectionName)
	endMarker := fmt.Sprintf(markerEndFmt, sectionName)

	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	existingStr := string(existing)
	startIdx := strings.Index(existingStr, startMarker)
	endIdx := strings.Index(existingStr, endMarker)
	if startIdx < 0 || endIdx <= startIdx {
		return nil
	}

	before := strings.TrimRight(existingStr[:startIdx], "\n")
	after := strings.TrimLeft(existingStr[endIdx+len(endMarker):], "\n")
	merged := before
	if len(after) > 0 {
		merged += "\n\n" + after
	} else {
		merged += "\n"
	}
	return writeFileEnsureDir(path, []byte(merged))
}

func mergeJSONFile(path string, ownedKeys map[string]any) error {
	existing, err := readJSONOrEmpty(path)
	if err != nil {
		return err
	}
	for k, v := range ownedKeys {
		// Deep-merge for "permissions" so we never clobber user customizations
		// (custom allow/deny entries, defaultMode preference, etc.).
		if k == "permissions" {
			tplPerms, tplOK := v.(map[string]any)
			userPerms, userOK := existing[k].(map[string]any)
			if tplOK && userOK {
				existing[k] = mergePermissions(userPerms, tplPerms)
				continue
			}
			// User had `permissions: null` or non-object → treat as missing,
			// fall through to the shallow assignment below (use template).
		}
		existing[k] = v
	}
	return writeJSON(path, existing)
}

// mergePermissions deep-merges the template's permissions block into the user's
// existing permissions block. Contract:
//   - allow/deny/ask: UNION (template entries first, then user-only entries),
//     deduped, stable order → idempotent across re-installs.
//   - defaultMode: preserve user's value if present; else use template's.
//   - any other sub-keys: prefer user value if present, else template value.
//
// existing is the user's current permissions object; template is what we ship.
// Both are non-nil maps (caller guarantees this).
func mergePermissions(existing, template map[string]any) map[string]any {
	out := map[string]any{}

	// 1. Start by copying every key the user already had so unknown sub-keys survive.
	for k, v := range existing {
		out[k] = v
	}

	// 2. List union for allow/deny/ask — only when the template provides the list.
	for _, listKey := range []string{"allow", "deny", "ask"} {
		tplList, tplOK := template[listKey].([]any)
		if !tplOK {
			// Template doesn't define this list → don't touch user's value (preserve as-is).
			continue
		}
		userList, _ := existing[listKey].([]any)
		out[listKey] = unionStringList(tplList, userList)
	}

	// 3. defaultMode: user wins if set; else template.
	if _, userHas := existing["defaultMode"]; !userHas {
		if v, ok := template["defaultMode"]; ok {
			out["defaultMode"] = v
		}
	}

	// 4. Any other template sub-keys not yet present in `out` get filled in from template.
	for k, v := range template {
		if k == "allow" || k == "deny" || k == "ask" || k == "defaultMode" {
			continue
		}
		if _, exists := out[k]; !exists {
			out[k] = v
		}
	}

	return out
}

// unionStringList returns the deduped union of two []any lists treated as strings.
// Template entries appear first in their original order, then user-only entries
// in their original order. Non-string elements are passed through (template-first).
func unionStringList(template, user []any) []any {
	seen := map[string]bool{}
	out := make([]any, 0, len(template)+len(user))

	for _, item := range template {
		s, ok := item.(string)
		if !ok {
			out = append(out, item)
			continue
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, item)
	}
	for _, item := range user {
		s, ok := item.(string)
		if !ok {
			out = append(out, item)
			continue
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, item)
	}
	return out
}

func mergeMCPServer(path, serverName string, config any) error {
	existing, err := readJSONOrEmpty(path)
	if err != nil {
		return err
	}
	mcpServers, ok := existing["mcpServers"].(map[string]any)
	if !ok {
		mcpServers = map[string]any{}
	}
	mcpServers[serverName] = config
	existing["mcpServers"] = mcpServers
	return writeJSON(path, existing)
}

func readJSONOrEmpty(path string) (map[string]any, error) {
	out := map[string]any{}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(raw) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}

func writeJSON(path string, data map[string]any) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return writeFileEnsureDir(path, append(out, '\n'))
}

func writeFileEnsureDir(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

