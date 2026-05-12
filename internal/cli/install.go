package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/lucianorepetti/system-general-ai/internal/claude"
	"github.com/lucianorepetti/system-general-ai/internal/engram"
	"github.com/lucianorepetti/system-general-ai/internal/gemini"
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

	var targets []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Target Platforms").
				Description("Choose the AI ecosystems to configure").
				Options(
					huh.NewOption("Claude Code", "claude").Selected(true),
					huh.NewOption("Gemini CLI", "gemini"),
				).
				Value(&targets),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("installation aborted: %w", err)
	}

	if len(targets) == 0 {
		return fmt.Errorf("no target platforms selected. Aborting")
	}

	out := cmd.OutOrStdout()

	for _, target := range targets {
		if target == "claude" {
			// 1. Detect Claude Code instances
			instances, err := claude.DetectInstances()
			if err != nil {
				return fmt.Errorf("detect Claude Code instances: %w", err)
			}
			if len(instances) == 0 {
				fmt.Fprintln(out, "Warning: No Claude Code installation found at ~/.claude or any .claude-work*.")
			} else {
				fmt.Fprintf(out, "\nFound %d Claude Code instance(s):\n", len(instances))
				for _, inst := range instances {
					fmt.Fprintf(out, "  • %s [%s]\n", inst.Path, claude.SourceLabel(inst))
				}
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
			if len(instances) > 0 {
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
			}
		} else if target == "gemini" {
			fmt.Fprintf(out, "\nSyncing Gemini CLI configuration ...\n")
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %w", err)
			}
			
			syncOpts := gemini.SyncOptions{
				Templates:    templates,
				EnableSDD:    true,
				EngramBinary: filepath.Join(defaultBinDir(), engramBinaryName()),
			}
			
			if err := gemini.Sync(cwd, syncOpts); err != nil {
				return fmt.Errorf("sync gemini: %w", err)
			}
			fmt.Fprintf(out, "  ✓ Configured GEMINI.md in %s\n", cwd)
		}
	}

	fmt.Fprintf(out, "\nDone. Restart Claude Code to apply.\n")
	return nil
}

func engramBinaryName() string {
	if runtime.GOOS == "windows" {
		return "engram.exe"
	}
	return "engram"
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

