package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lucianorepetti/system-general-ai/internal/claude"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-apply templates and settings to every detected Claude Code instance",
	Long: `Use this after installing a new Claude Code instance, after upgrading
system-general-ai, or to recover from manual edits that broke the configuration.

Sync is idempotent: re-running with the same defaults produces the same result.`,
	RunE: runSync,
}

func runSync(cmd *cobra.Command, _ []string) error {
	instances, err := claude.DetectInstances()
	if err != nil {
		return fmt.Errorf("detect instances: %w", err)
	}
	if len(instances) == 0 {
		return fmt.Errorf("no Claude Code installation found. Run `system-general-ai install` first")
	}

	binDir := defaultBinDir()
	engramPath := filepath.Join(binDir, engramBinaryName())

	syncOpts := claude.SyncOptions{
		Templates:    templates,
		Mode:         claude.ModePermissive,
		EnableSDD:    true,
		EngramBinary: engramPath,
	}

	out := cmd.OutOrStdout()
	for _, inst := range instances {
		if err := claude.Sync(inst, syncOpts); err != nil {
			return fmt.Errorf("sync %s: %w", inst.Path, err)
		}
		fmt.Fprintf(out, "  ✓ %s\n", inst.Path)
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

