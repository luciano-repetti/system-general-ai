package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/lucianorepetti/system-general-ai/internal/claude"
	codexpkg "github.com/lucianorepetti/system-general-ai/internal/codex"
	"github.com/lucianorepetti/system-general-ai/internal/gemini"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove every system-general-ai-managed artifact from each detected instance",
	Long: `Removes the artifacts owned by system-general-ai. User content is preserved
when the wrapper can recognize what's ours vs theirs.

Specifically:
  - The system-general-ai output-style file
  - The marker-delimited sections in CLAUDE.md / GEMINI.md
  - The system-general-ai skill files
  - settings.json (Claude): outputStyle, permissions, mcpServers.engram
  - Permissive policy (Gemini)
  - Engram binary (only with --remove-engram)

Note: settings.json values that we previously OVERWROTE (e.g. a non-default
outputStyle, custom permissions) cannot be restored from the uninstall step
alone — re-apply your previous configuration afterwards if needed.`,
	RunE: runUninstall,
}

var removeEngram bool
var uninstallTargets []string
var uninstallAll bool

func init() {
	uninstallCmd.Flags().BoolVar(&removeEngram, "remove-engram", false, "also remove the engram binary from PATH")
	uninstallCmd.Flags().StringSliceVar(&uninstallTargets, "target", nil, "target platform to uninstall (claude, gemini, codex); repeatable or comma-separated")
	uninstallCmd.Flags().BoolVar(&uninstallAll, "all", false, "uninstall all target platforms without opening the selector")
}

func runUninstall(cmd *cobra.Command, _ []string) error {
	targets, err := normalizeTargets(uninstallTargets, uninstallAll, nil)
	if err != nil {
		if len(uninstallTargets) > 0 || uninstallAll {
			return err
		}
		var selected []string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select Platforms to Uninstall").
					Description("Choose the AI ecosystems to clean up").
					Options(
						huh.NewOption("Claude Code", "claude").Selected(true),
						huh.NewOption("Gemini CLI", "gemini").Selected(true),
						huh.NewOption("Codex", "codex").Selected(true),
					).
					Value(&selected),
			),
		)

		if formErr := form.Run(); formErr != nil {
			return fmt.Errorf("uninstallation aborted: %w", formErr)
		}
		targets, err = normalizeTargets(selected, false, nil)
		if err != nil {
			return err
		}
	}

	out := cmd.OutOrStdout()

	for _, target := range targets {
		if target == "claude" {
			instances, err := claude.DetectInstances()
			if err != nil {
				return err
			}
			for _, inst := range instances {
				if err := uninstallFromInstance(inst); err != nil {
					fmt.Fprintf(out, "  ✗ Claude: %s — %v\n", inst.Path, err)
					continue
				}
				fmt.Fprintf(out, "  ✓ Claude: %s\n", inst.Path)
			}
		} else if target == "gemini" {
			if err := uninstallFromGemini(); err != nil {
				fmt.Fprintf(out, "  ✗ Gemini: %v\n", err)
			} else {
				fmt.Fprintf(out, "  ✓ Gemini: successfully cleaned up\n")
			}
		} else if target == "codex" {
			inst, err := codexpkg.DetectInstance()
			if err != nil {
				fmt.Fprintf(out, "  ✗ Codex: %v\n", err)
				continue
			}
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(out, "  ✗ Codex: %v\n", err)
				continue
			}
			if err := codexpkg.Uninstall(inst, cwd, false); err != nil {
				fmt.Fprintf(out, "  ✗ Codex: %s — %v\n", inst.Path, err)
			} else {
				fmt.Fprintf(out, "  ✓ Codex: %s\n", inst.Path)
			}
		}
	}

	if removeEngram {
		path := filepath.Join(defaultBinDir(), engramBinaryName())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(out, "  ✗ remove engram: %v\n", err)
		} else {
			fmt.Fprintf(out, "  ✓ removed %s\n", path)
		}
	}

	fmt.Fprintln(out, "\nDone.")
	return nil
}

func uninstallFromGemini() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 1. Remove marker sections from GEMINI.md
	geminiPath := filepath.Join(cwd, "GEMINI.md")
	for _, section := range []string{"global-rules", "sdd-orchestrator"} {
		if err := gemini.RemoveMarkedSection(geminiPath, section); err != nil {
			return fmt.Errorf("remove gemini marker %s: %w", section, err)
		}
	}

	// 2. Remove our skills from ~/.gemini/skills/
	home, _ := os.UserHomeDir()
	managed := []string{
		"sdd-init", "sdd-explore", "sdd-propose",
		"sdd-spec", "sdd-design", "sdd-tasks",
		"sdd-apply", "sdd-verify", "sdd-archive",
		"shell-runner", "compact-suggest",
	}
	for _, name := range managed {
		skillDir := filepath.Join(home, ".gemini", "skills", name)
		_ = os.Remove(filepath.Join(skillDir, "SKILL.md"))
		_ = os.Remove(skillDir)
	}

	// 3. Remove permissive policy
	policyPath := filepath.Join(home, ".gemini", "policies", "permissive.toml")
	if err := os.Remove(policyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove gemini policy: %w", err)
	}

	return nil
}

func uninstallFromInstance(inst claude.Instance) error {
	// 1. Remove output-style file
	osPath := filepath.Join(inst.OutputStylesDir(), "system-general-ai.md")
	if err := os.Remove(osPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove output style: %w", err)
	}

	// 2. Remove marker sections from CLAUDE.md
	for _, section := range []string{"global-rules", "sdd-orchestrator"} {
		if err := claude.RemoveMarkedSection(inst.CLAUDEMDPath(), section); err != nil {
			return fmt.Errorf("remove marker %s: %w", section, err)
		}
	}

	// 3. Remove our skills (best-effort — never touch user-authored skills).
	// Skills are installed as directories: <skills>/<name>/SKILL.md
	managed := []string{
		"sdd-init", "sdd-explore", "sdd-propose",
		"sdd-spec", "sdd-design", "sdd-tasks",
		"sdd-apply", "sdd-verify", "sdd-archive",
		"shell-runner", "compact-suggest",
	}
	for _, name := range managed {
		skillDir := filepath.Join(inst.SkillsDir(), name)
		// Remove SKILL.md first, then the dir if empty (preserves any user-authored
		// extras like notes.md or images dropped under the same skill folder).
		_ = os.Remove(filepath.Join(skillDir, "SKILL.md"))
		_ = os.Remove(skillDir) // succeeds only if dir is empty
	}

	// 4. Clean settings.json — only keys we recognize as ours
	if err := cleanOurSettings(inst.SettingsPath()); err != nil {
		return fmt.Errorf("clean settings: %w", err)
	}

	return nil
}

// cleanOurSettings removes keys/subkeys that we recognize as written by us.
// User-authored keys are preserved.
func cleanOurSettings(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(raw) == 0 {
		return nil
	}

	var settings map[string]any
	if err := json.Unmarshal(raw, &settings); err != nil {
		return err
	}

	// outputStyle: remove only if it points at us
	if v, ok := settings["outputStyle"].(string); ok && v == "system-general-ai" {
		delete(settings, "outputStyle")
	}

	// mcpServers.engram: remove only if command points at OUR binary path
	if mcps, ok := settings["mcpServers"].(map[string]any); ok {
		if engram, ok := mcps["engram"].(map[string]any); ok {
			if cmd, ok := engram["command"].(string); ok {
				ourBinDir := defaultBinDir()
				if strings.Contains(strings.ToLower(cmd), strings.ToLower(ourBinDir)) {
					delete(mcps, "engram")
					if len(mcps) == 0 {
						delete(settings, "mcpServers")
					} else {
						settings["mcpServers"] = mcps
					}
				}
			}
		}
	}

	// permissions: remove only if its allow-list shape matches one of our presets
	if perms, ok := settings["permissions"].(map[string]any); ok && looksLikeOurPermissions(perms) {
		delete(settings, "permissions")
	}

	// $schema: harmless to leave; remove only if we set it (heuristic: any json.schemastore.org)
	if v, ok := settings["$schema"].(string); ok && strings.Contains(v, "json.schemastore.org/claude-code-settings") {
		// Only drop if everything else of ours is also gone — user may have legitimately set it
		// To stay conservative, leave it.
		_ = v
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// looksLikeOurPermissions returns true when the permissions block looks like one
// of the three presets we ship (permissive/balanced/strict). Heuristic: presence
// of our characteristic deny-list entries.
func looksLikeOurPermissions(perms map[string]any) bool {
	deny, ok := perms["deny"].([]any)
	if !ok {
		return false
	}
	hits := 0
	signatures := []string{
		"Bash(rm -rf /*)",
		"Bash(git push --force*)",
		"Bash(git reset --hard*)",
	}
	for _, item := range deny {
		s, _ := item.(string)
		for _, sig := range signatures {
			if s == sig {
				hits++
			}
		}
	}
	return hits >= 2
}
