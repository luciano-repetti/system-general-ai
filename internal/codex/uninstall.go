package codex

import (
	"fmt"
	"os"
	"path/filepath"
)

// Uninstall removes system-general-ai-managed Codex artifacts.
func Uninstall(inst Instance, projectDir string, removeEngram bool) error {
	if projectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectDir = cwd
	}

	if err := cleanAGENTS(inst.AGENTSPath(), []string{"global-rules", "output-style", "sdd-orchestrator", "persona"}); err != nil {
		return err
	}
	if err := cleanAGENTS(filepath.Join(projectDir, "AGENTS.md"), []string{"project-rules"}); err != nil {
		return err
	}
	if err := cleanConfig(inst.ConfigPath()); err != nil {
		return err
	}

	for _, name := range managedSkills() {
		skillDir := filepath.Join(inst.SkillsDir(), name)
		_ = os.Remove(filepath.Join(skillDir, "SKILL.md"))
		_ = os.Remove(skillDir)
	}
	_ = os.Remove(filepath.Join(inst.SkillsDir(), "_shared", "SKILL.md"))
	_ = os.Remove(filepath.Join(inst.SkillsDir(), "_shared", "engram-convention.md"))
	_ = os.Remove(filepath.Join(inst.SkillsDir(), "_shared", "sdd-phase-common.md"))
	_ = os.Remove(filepath.Join(inst.SkillsDir(), "_shared"))

	_ = os.Remove(inst.EngramInstructionsPath())
	_ = os.Remove(inst.EngramCompactPromptPath())

	if removeEngram {
		// The caller owns actual Engram binary removal because it is shared
		// across Claude, Gemini, and Codex targets.
	}
	return nil
}

func cleanAGENTS(path string, sections []string) error {
	content, err := readTextOrEmpty(path)
	if err != nil {
		return err
	}
	updated := content
	for _, section := range sections {
		updated = removeMarkedSection(updated, section)
	}
	if updated == content {
		return nil
	}
	if updated == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
		return nil
	}
	return writeTextEnsureDir(path, updated)
}

func cleanConfig(path string) error {
	content, err := readTextOrEmpty(path)
	if err != nil {
		return err
	}
	updated := removeCodexEngramBlock(content)
	updated = removeTopLevelTOMLKey(updated, "model_instructions_file")
	updated = removeTopLevelTOMLKey(updated, "experimental_compact_prompt_file")
	if updated == content {
		return nil
	}
	if updated == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return writeTextEnsureDir(path, updated)
}

func managedSkills() []string {
	return []string{
		"sdd-init", "sdd-explore", "sdd-propose",
		"sdd-spec", "sdd-design", "sdd-tasks",
		"sdd-apply", "sdd-verify", "sdd-archive",
		"shell-runner", "compact-suggest",
	}
}
