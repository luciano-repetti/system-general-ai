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

	pathsToBackup := []string{
		inst.ConfigPath(),
		inst.AGENTSPath(),
		filepath.Join(opts.ProjectDir, "AGENTS.md"),
	}
	if err := backupExisting(inst, pathsToBackup); err != nil {
		return fmt.Errorf("backup codex config: %w", err)
	}

	if err := syncGlobalAGENTS(inst, opts); err != nil {
		return fmt.Errorf("global AGENTS.md: %w", err)
	}
	if err := syncProjectAGENTS(opts); err != nil {
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

	return writeTextEnsureDir(inst.AGENTSPath(), existing)
}

func syncProjectAGENTS(opts SyncOptions) error {
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
	return writeTextEnsureDir(projectPath, existing)
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

	updated := upsertTopLevelTOMLString(existing, "model_instructions_file", inst.EngramInstructionsPath())
	updated = upsertTopLevelTOMLString(updated, "experimental_compact_prompt_file", inst.EngramCompactPromptPath())
	if opts.EngramBinary != "" {
		updated = upsertCodexEngramBlock(updated, opts.EngramBinary)
	}

	return writeTextEnsureDir(inst.ConfigPath(), updated)
}

func backupExisting(inst Instance, paths []string) error {
	stamp := time.Now().UTC().Format("20060102-150405")
	backupDir := filepath.Join(inst.Path, "backups", "system-general-ai-"+stamp)
	wrote := false

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		rel := strings.TrimPrefix(filepath.Clean(p), filepath.VolumeName(filepath.Clean(p)))
		rel = strings.TrimLeft(rel, string(filepath.Separator))
		dst := filepath.Join(backupDir, rel)
		if err := copyFilePath(p, dst); err != nil {
			return err
		}
		wrote = true
	}

	if !wrote {
		return nil
	}
	return nil
}

func copyFromFS(srcFS fs.FS, srcPath, dst string) error {
	in, err := srcFS.Open(path.Clean(srcPath))
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
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
