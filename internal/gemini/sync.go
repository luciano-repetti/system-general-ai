package gemini

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SyncOptions contains configuration for syncing Gemini CLI templates
type SyncOptions struct {
	Templates fs.FS
	EnableSDD bool
}

const (
	markerStartFmt = "<!-- BEGIN: system-general-ai/%s -->"
	markerEndFmt   = "<!-- END: system-general-ai/%s -->"
)

// Sync applies the Gemini CLI configuration to the specified project directory
// and the global Gemini CLI configuration directory.
func Sync(projectDir string, opts SyncOptions) error {
	// 1. Sync GEMINI.md in the project directory
	if err := syncProjectRules(projectDir, opts); err != nil {
		return err
	}

	// 2. Sync skills to global Gemini CLI config (~/.gemini/skills/)
	if err := syncGlobalSkills(opts); err != nil {
		return err
	}

	// 3. Sync permissive policy to global Gemini CLI config (~/.gemini/policies/)
	if err := syncGlobalPolicy(); err != nil {
		return err
	}

	return nil
}

func syncProjectRules(projectDir string, opts SyncOptions) error {
	path := filepath.Join(projectDir, "GEMINI.md")
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read GEMINI.md: %w", err)
	}

	// Sync global rules
	globalRules, err := fs.ReadFile(opts.Templates, "templates/gemini/GEMINI.md.global.tmpl")
	if err != nil {
		return fmt.Errorf("read global rules template: %w", err)
	}
	content = mergeMarkedSection(content, "global-rules", globalRules)

	// Sync SDD orchestrator if enabled
	if opts.EnableSDD {
		orchestrator, err := fs.ReadFile(opts.Templates, "templates/gemini/orchestrator.md")
		if err != nil {
			return fmt.Errorf("read orchestrator template: %w", err)
		}
		content = mergeMarkedSection(content, "sdd-orchestrator", orchestrator)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("write GEMINI.md: %w", err)
	}
	return nil
}

func syncGlobalSkills(opts SyncOptions) error {
	home, _ := os.UserHomeDir()
	globalSkillsDir := filepath.Join(home, ".gemini", "skills")

	return fs.WalkDir(opts.Templates, "templates/gemini/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// Calculate relative path from templates/gemini/skills
		rel, _ := filepath.Rel("templates/gemini/skills", path)
		dst := filepath.Join(globalSkillsDir, rel)

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}

		data, _ := fs.ReadFile(opts.Templates, path)
		return os.WriteFile(dst, data, 0644)
	})
}

func syncGlobalPolicy() error {
	home, _ := os.UserHomeDir()
	policyPath := filepath.Join(home, ".gemini", "policies", "permissive.toml")
	
	policy := `[[rule]]
toolName = "*"
decision = "allow"
priority = 999

[[rule]]
toolName = "run_shell_command"
decision = "allow"
priority = 1000
allowRedirection = true`

	if err := os.MkdirAll(filepath.Dir(policyPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(policyPath, []byte(policy), 0644)
}

func RemoveMarkedSection(path string, id string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	startMarker := fmt.Sprintf(markerStartFmt, id)
	endMarker := fmt.Sprintf(markerEndFmt, id)

	startIdx := strings.Index(string(content), startMarker)
	endIdx := strings.Index(string(content), endMarker)

	if startIdx == -1 || endIdx == -1 {
		return nil
	}

	// Remove section and the following newline if any
	result := append(content[:startIdx], content[endIdx+len(endMarker):]...)
	
	// Trim leading/trailing whitespace around the cut
	trimmed := strings.TrimSpace(string(result))
	if trimmed == "" {
		return os.Remove(path)
	}
	
	return os.WriteFile(path, []byte(trimmed+"\n"), 0644)
}

func mergeMarkedSection(content []byte, id string, section []byte) []byte {
	startMarker := fmt.Sprintf(markerStartFmt, id)
	endMarker := fmt.Sprintf(markerEndFmt, id)
	
	newSection := bytes.Join([][]byte{
		[]byte(startMarker),
		bytes.TrimSpace(section),
		[]byte(endMarker),
	}, []byte("\n"))

	startIdx := strings.Index(string(content), startMarker)
	endIdx := strings.Index(string(content), endMarker)

	if startIdx == -1 || endIdx == -1 {
		// Append to end if not found
		if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
			content = append(content, '\n')
		}
		return append(content, append(newSection, '\n')...)
	}

	// Replace existing section
	result := make([]byte, 0, len(content)- (endIdx+len(endMarker)-startIdx) + len(newSection))
	result = append(result, content[:startIdx]...)
	result = append(result, newSection...)
	result = append(result, content[endIdx+len(endMarker):]...)
	return result
}
