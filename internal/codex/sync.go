package codex

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// SyncOptions controls what gets pushed to a Codex config root.
type SyncOptions struct {
	Templates    fs.FS
	ProjectDir   string
	EnableSDD    bool
	EngramBinary string
	Persona      string
}

// Sync applies system-general-ai to Codex.
func Sync(inst Instance, opts SyncOptions) error {
	if opts.ProjectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve project dir: %w", err)
		}
		opts.ProjectDir = cwd
	}

	if err := ensureDirs(inst); err != nil {
		return err
	}
	if err := validateExistingCodexConfig(inst); err != nil {
		return err
	}

	if err := syncGlobalAGENTS(inst, opts); err != nil {
		return fmt.Errorf("global AGENTS.md: %w", err)
	}
	if err := syncProjectAGENTS(inst, opts); err != nil {
		return fmt.Errorf("project AGENTS.md: %w", err)
	}
	if err := syncSkills(inst, opts); err != nil {
		return fmt.Errorf("skills: %w", err)
	}
	if err := syncEngramFiles(inst, opts); err != nil {
		return fmt.Errorf("engram instruction files: %w", err)
	}
	if err := syncConfig(inst, opts); err != nil {
		return fmt.Errorf("config.toml: %w", err)
	}

	return nil
}

func validateExistingCodexConfig(inst Instance) error {
	existing, err := readTextOrEmpty(inst.ConfigPath())
	if err != nil {
		return err
	}
	if err := validateTOML(existing); err != nil {
		return fmt.Errorf("existing config.toml is invalid TOML: %w", err)
	}
	return nil
}

func ensureDirs(inst Instance) error {
	for _, dir := range []string{inst.Path, inst.SkillsDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	return nil
}

func syncGlobalAGENTS(inst Instance, opts SyncOptions) error {
	existing, err := readTextOrEmpty(inst.AGENTSPath())
	if err != nil {
		return err
	}

	globalRules, err := fs.ReadFile(opts.Templates, "codex/AGENTS.md.global.tmpl")
	if err != nil {
		return fmt.Errorf("read codex global template: %w", err)
	}
	existing = mergeMarkedSection(existing, "global-rules", string(globalRules))

	outputStyle, err := fs.ReadFile(opts.Templates, "output-style.md")
	if err != nil {
		return fmt.Errorf("read output style: %w", err)
	}
	existing = mergeMarkedSection(existing, "output-style", stripFrontmatter(string(outputStyle)))

	if opts.EnableSDD {
		orchestrator, err := fs.ReadFile(opts.Templates, "codex/orchestrator.md")
		if err != nil {
			return fmt.Errorf("read codex orchestrator: %w", err)
		}
		existing = mergeMarkedSection(existing, "sdd-orchestrator", string(orchestrator))
	} else {
		existing = removeMarkedSection(existing, "sdd-orchestrator")
	}

	if strings.TrimSpace(opts.Persona) != "" {
		persona := "## Persona\n\n" + strings.TrimSpace(opts.Persona) + "\n"
		existing = mergeMarkedSection(existing, "persona", persona)
	} else {
		existing = removeMarkedSection(existing, "persona")
	}

	return writeManagedText(inst, inst.AGENTSPath(), existing)
}

func syncProjectAGENTS(inst Instance, opts SyncOptions) error {
	projectPath := filepath.Join(opts.ProjectDir, "AGENTS.md")
	existing, err := readTextOrEmpty(projectPath)
	if err != nil {
		return err
	}

	projectRules, err := fs.ReadFile(opts.Templates, "codex/AGENTS.md.project.tmpl")
	if err != nil {
		return fmt.Errorf("read codex project template: %w", err)
	}
	existing = mergeMarkedSection(existing, "project-rules", string(projectRules))
	return writeManagedText(inst, projectPath, existing)
}

func syncSkills(inst Instance, opts SyncOptions) error {
	for _, root := range []string{"skills", "codex/skills"} {
		if err := fs.WalkDir(opts.Templates, root, func(src string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(filepath.FromSlash(root), filepath.FromSlash(src))
			if err != nil {
				return err
			}
			dst := filepath.Join(inst.SkillsDir(), rel)
			return copyFromFS(opts.Templates, src, dst)
		}); err != nil {
			return err
		}
	}
	return nil
}

func syncEngramFiles(inst Instance, opts SyncOptions) error {
	for _, item := range []struct {
		src string
		dst string
	}{
		{"codex/engram-instructions.md", inst.EngramInstructionsPath()},
		{"codex/engram-compact-prompt.md", inst.EngramCompactPromptPath()},
	} {
		if err := copyFromFS(opts.Templates, item.src, item.dst); err != nil {
			return fmt.Errorf("copy %s: %w", item.src, err)
		}
	}
	return nil
}

func syncConfig(inst Instance, opts SyncOptions) error {
	existing, err := readTextOrEmpty(inst.ConfigPath())
	if err != nil {
		return err
	}
	if err := validateTOML(existing); err != nil {
		return fmt.Errorf("existing config.toml is invalid TOML: %w", err)
	}

	updated := upsertTopLevelTOMLStrings(existing, []tomlStringKV{
		{key: "model_instructions_file", value: inst.EngramInstructionsPath()},
		{key: "experimental_compact_prompt_file", value: inst.EngramCompactPromptPath()},
	})
	if opts.EngramBinary != "" {
		updated = upsertCodexEngramBlock(updated, opts.EngramBinary)
	}
	if err := validateTOML(updated); err != nil {
		return fmt.Errorf("generated config.toml is invalid TOML: %w", err)
	}

	return writeManagedText(inst, inst.ConfigPath(), updated)
}

func backupExisting(inst Instance, paths []string) error {
	for _, p := range paths {
		if err := backupExistingFile(inst, p); err != nil {
			return err
		}
	}
	return nil
}

func backupExistingFile(inst Instance, p string) error {
	stamp := time.Now().UTC().Format("20060102-150405")
	backupDir := filepath.Join(inst.Path, "backups", "system-general-ai-"+stamp)

	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	rel := strings.TrimPrefix(filepath.Clean(p), filepath.VolumeName(filepath.Clean(p)))
	rel = strings.TrimLeft(rel, string(filepath.Separator))
	dst := filepath.Join(backupDir, rel)
	return copyFilePath(p, dst)
}

func copyFromFS(srcFS fs.FS, srcPath, dst string) error {
	raw, err := fs.ReadFile(srcFS, path.Clean(srcPath))
	if err != nil {
		return err
	}
	return writeBytesEnsureDirIfChanged(dst, raw)
}

func copyFilePath(src, dst string) error {
	in, err := os.Open(src)
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

func readTextOrEmpty(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(raw), nil
}

func writeTextEnsureDir(path, content string) error {
	return writeBytesEnsureDirIfChanged(path, []byte(normalizeFinalNewline(content)))
}

func writeManagedText(inst Instance, path, content string) error {
	normalized := normalizeFinalNewline(content)
	existing, err := os.ReadFile(path)
	if err == nil && string(existing) == normalized {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err == nil {
		if info, statErr := os.Stat(path); statErr == nil && info.Mode().IsRegular() {
			if err := backupExistingFile(inst, path); err != nil {
				return fmt.Errorf("backup %s: %w", path, err)
			}
		}
	}
	return writeBytesEnsureDirIfChanged(path, []byte(normalized))
}

func writeBytesEnsureDirIfChanged(path string, content []byte) error {
	existing, err := os.ReadFile(path)
	if err == nil && string(existing) == string(content) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func normalizeFinalNewline(content string) string {
	if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content
}

func stripFrontmatter(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return content
	}
	end := strings.Index(normalized[4:], "\n---\n")
	if end < 0 {
		return content
	}
	return strings.TrimLeft(normalized[4+end+5:], "\n")
}
