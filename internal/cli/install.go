package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/lucianorepetti/system-general-ai/internal/claude"
	"github.com/lucianorepetti/system-general-ai/internal/engram"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install system-general-ai into every detected Claude Code instance (zero-config)",
	Long: `Detects all Claude Code installations on this machine, installs the
engram binary to a user-level PATH, and synchronizes templates,
settings and MCP configuration.

Defaults are applied — no questions asked. Reconfigure later with:
  system-general-ai configure permissions <permissive|balanced|strict>
  system-general-ai configure persona <name>`,
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	// 1. Detect Claude Code instances
	instances, err := claude.DetectInstances()
	if err != nil {
		return fmt.Errorf("detect Claude Code instances: %w", err)
	}
	if len(instances) == 0 {
		return fmt.Errorf("no Claude Code installation found at ~/.claude or any .claude-work*. Install Claude Code first")
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Found %d Claude Code instance(s):\n", len(instances))
	for _, inst := range instances {
		fmt.Fprintf(out, "  • %s [%s]\n", inst.Path, claude.SourceLabel(inst))
	}

	// 2. Install engram binary to user-level PATH
	binDir := defaultBinDir()
	fmt.Fprintf(out, "\nInstalling engram to %s ...\n", binDir)
	engramPath, err := engram.Install(ctx, engram.InstallOptions{TargetDir: binDir})
	if err != nil {
		return fmt.Errorf("install engram: %w", err)
	}
	fmt.Fprintf(out, "  ✓ %s\n", engramPath)

	// 3. Apply templates + settings + MCP to each instance (zero-config defaults)
	fmt.Fprintf(out, "\nSyncing configuration to instances ...\n")
	syncOpts := claude.SyncOptions{
		Templates:    templates,
		Mode:         claude.ModePermissive,
		EnableSDD:    true,
		EngramBinary: engramPath,
	}
	for _, inst := range instances {
		if err := claude.Sync(inst, syncOpts); err != nil {
			return fmt.Errorf("sync %s: %w", inst.Path, err)
		}
		fmt.Fprintf(out, "  ✓ %s\n", inst.Path)
	}

	fmt.Fprintf(out, "\nDone. Restart Claude Code to apply.\n")
	return nil
}

// defaultBinDir returns the user-level bin directory we install engram into.
//   Linux/macOS: $HOME/.local/bin
//   Windows:     %LOCALAPPDATA%\system-general-ai\bin
func defaultBinDir() string {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("LOCALAPPDATA")
		if appdata == "" {
			appdata = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(appdata, "system-general-ai", "bin")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "bin")
}

