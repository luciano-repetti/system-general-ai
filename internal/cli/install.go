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
	codexpkg "github.com/lucianorepetti/system-general-ai/internal/codex"
	"github.com/lucianorepetti/system-general-ai/internal/engram"
	"github.com/lucianorepetti/system-general-ai/internal/gemini"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install system-general-ai into Claude Code, Gemini CLI, and Codex",
	Long: `Detects selected AI ecosystems, installs the engram binary to a
user-level PATH, and synchronizes templates, settings and MCP configuration.

Defaults are applied. Reconfigure later with:
  system-general-ai configure permissions <permissive|balanced|strict>
  system-general-ai configure persona <name>`,
	RunE: runInstall,
}

var installTargets []string
var installAll bool

func init() {
	installCmd.Flags().StringSliceVar(&installTargets, "target", nil, "target platform to install (claude, gemini, codex); repeatable or comma-separated")
	installCmd.Flags().BoolVar(&installAll, "all", false, "install all target platforms without opening the selector")
}

func runInstall(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	targets, err := normalizeTargets(installTargets, installAll, nil)
	if err != nil {
		if len(installTargets) > 0 || installAll {
			return err
		}
		var selected []string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select Target Platforms").
					Description("Choose the AI ecosystems to configure").
					Options(
						huh.NewOption("Claude Code", "claude").Selected(true),
						huh.NewOption("Gemini CLI", "gemini"),
						huh.NewOption("Codex", "codex").Selected(true),
					).
					Value(&selected),
			),
		)

		if formErr := form.Run(); formErr != nil {
			return fmt.Errorf("installation aborted: %w", formErr)
		}
		targets, err = normalizeTargets(selected, false, nil)
		if err != nil {
			return err
		}
	}

	out := cmd.OutOrStdout()

	binDir := defaultBinDir()
	fmt.Fprintf(out, "\nInstalling engram to %s ...\n", binDir)
	engramPath, err := engram.Install(ctx, engram.InstallOptions{TargetDir: binDir})
	if err != nil {
		return fmt.Errorf("install engram: %w", err)
	}
	fmt.Fprintf(out, "  done: %s\n", engramPath)

	for _, target := range targets {
		switch target {
		case "claude":
			if err := installClaude(out, engramPath); err != nil {
				return err
			}
		case "gemini":
			if err := installGemini(out, engramPath); err != nil {
				return err
			}
		case "codex":
			if err := installCodex(out, engramPath); err != nil {
				return err
			}
		}
	}

	fmt.Fprintf(out, "\nDone. Restart configured AI tools to apply.\n")
	return nil
}

func installClaude(out interface {
	Write([]byte) (int, error)
}, engramPath string) error {
	instances, err := claude.DetectInstances()
	if err != nil {
		return fmt.Errorf("detect Claude Code instances: %w", err)
	}
	if len(instances) == 0 {
		fmt.Fprintln(out, "Warning: No Claude Code installation found at ~/.claude or any .claude-work*.")
		return nil
	}

	fmt.Fprintf(out, "\nFound %d Claude Code instance(s):\n", len(instances))
	for _, inst := range instances {
		fmt.Fprintf(out, "  - %s [%s]\n", inst.Path, claude.SourceLabel(inst))
	}

	fmt.Fprintf(out, "\nSyncing Claude Code configuration ...\n")
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
		fmt.Fprintf(out, "  done: %s\n", inst.Path)
	}
	return nil
}

func installGemini(out interface {
	Write([]byte) (int, error)
}, engramPath string) error {
	fmt.Fprintf(out, "\nSyncing Gemini CLI configuration ...\n")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if err := gemini.Sync(cwd, gemini.SyncOptions{
		Templates:    templates,
		EnableSDD:    true,
		EngramBinary: engramPath,
	}); err != nil {
		return fmt.Errorf("sync gemini: %w", err)
	}
	fmt.Fprintf(out, "  done: GEMINI.md in %s\n", cwd)
	return nil
}

func installCodex(out interface {
	Write([]byte) (int, error)
}, engramPath string) error {
	fmt.Fprintf(out, "\nSyncing Codex configuration ...\n")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	inst, err := codexpkg.DetectInstance()
	if err != nil {
		return fmt.Errorf("detect Codex: %w", err)
	}
	if err := codexpkg.Sync(inst, codexpkg.SyncOptions{
		Templates:    templates,
		ProjectDir:   cwd,
		EnableSDD:    true,
		EngramBinary: engramPath,
	}); err != nil {
		return fmt.Errorf("sync codex: %w", err)
	}
	fmt.Fprintf(out, "  done: %s\n", inst.Path)
	return nil
}

// defaultBinDir returns the user-level bin directory we install engram into.
//
//	Linux/macOS: $HOME/.local/bin
//	Windows:     %LOCALAPPDATA%\system-general-ai\bin
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
