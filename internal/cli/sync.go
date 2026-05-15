package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lucianorepetti/system-general-ai/internal/claude"
	codexpkg "github.com/lucianorepetti/system-general-ai/internal/codex"
	"github.com/lucianorepetti/system-general-ai/internal/gemini"
)

var syncCmd = &cobra.Command{
	Use:   "sync [claude|gemini|codex...]",
	Short: "Re-apply templates and settings to detected AI tools",
	Long: `Use this after installing a new AI tool instance, after upgrading
system-general-ai, or to recover from manual edits that broke the configuration.

Sync is idempotent: re-running with the same defaults produces the same result.`,
	RunE: runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	binDir := defaultBinDir()
	engramPath := filepath.Join(binDir, engramBinaryName())
	out := cmd.OutOrStdout()

	targets, err := normalizeTargets(args, false, allTargets)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd for sync: %w", err)
	}

	for _, target := range targets {
		switch target {
		case "claude":
			if err := syncClaude(out, engramPath); err != nil {
				return err
			}
		case "gemini":
			if err := gemini.Sync(cwd, gemini.SyncOptions{
				Templates:    templates,
				EnableSDD:    true,
				EngramBinary: engramPath,
			}); err != nil {
				return fmt.Errorf("sync gemini: %w", err)
			}
			fmt.Fprintf(out, "  done: Gemini %s\n", cwd)
		case "codex":
			codexInst, err := codexpkg.DetectInstance()
			if err != nil {
				return fmt.Errorf("detect Codex: %w", err)
			}
			if err := codexpkg.Sync(codexInst, codexpkg.SyncOptions{
				Templates:    templates,
				ProjectDir:   cwd,
				EnableSDD:    true,
				EngramBinary: engramPath,
			}); err != nil {
				return fmt.Errorf("sync codex %s: %w", codexInst.Path, err)
			}
			fmt.Fprintf(out, "  done: Codex %s\n", codexInst.Path)
		}
	}
	return nil
}

func syncClaude(out interface {
	Write([]byte) (int, error)
}, engramPath string) error {
	instances, err := claude.DetectInstances()
	if err != nil {
		return fmt.Errorf("detect Claude instances: %w", err)
	}
	if len(instances) == 0 {
		fmt.Fprintln(out, "Warning: no Claude Code installation found.")
		return nil
	}
	for _, inst := range instances {
		if err := claude.Sync(inst, claude.SyncOptions{
			Templates:    templates,
			Mode:         claude.ModePermissive,
			EnableSDD:    true,
			EngramBinary: engramPath,
		}); err != nil {
			return fmt.Errorf("sync %s: %w", inst.Path, err)
		}
		fmt.Fprintf(out, "  done: Claude %s\n", inst.Path)
	}
	return nil
}

// engramBinaryName returns the platform-appropriate engram filename.
func engramBinaryName() string {
	if isWindows() {
		return "engram.exe"
	}
	return "engram"
}
