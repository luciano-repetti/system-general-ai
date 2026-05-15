package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/lucianorepetti/system-general-ai/internal/claude"
	codexpkg "github.com/lucianorepetti/system-general-ai/internal/codex"
	"github.com/lucianorepetti/system-general-ai/internal/gemini"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Change post-install settings (permissions, persona)",
}

var configurePermissionsCmd = &cobra.Command{
	Use:   "permissions <permissive|balanced|strict>",
	Short: "Switch the permission profile for every detected Claude Code instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigurePermissions,
}

var configurePersonaCmd = &cobra.Command{
	Use:   "persona <name>",
	Short: "Switch the active persona. Pass an empty string to remove it (use \"-\" or \"neutral\")",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigurePersona,
}

var configurePersonaTargets []string
var configurePersonaAll bool

func init() {
	configurePersonaCmd.Flags().StringSliceVar(&configurePersonaTargets, "target", nil, "target platform to configure (claude, gemini, codex); repeatable or comma-separated")
	configurePersonaCmd.Flags().BoolVar(&configurePersonaAll, "all", false, "configure all target platforms")
	configureCmd.AddCommand(configurePermissionsCmd)
	configureCmd.AddCommand(configurePersonaCmd)
}

func runConfigurePermissions(cmd *cobra.Command, args []string) error {
	mode := claude.PermissionMode(args[0])
	switch mode {
	case claude.ModePermissive, claude.ModeBalanced, claude.ModeStrict:
	default:
		return fmt.Errorf("unknown mode %q; expected one of: permissive, balanced, strict", args[0])
	}

	instances, err := claude.DetectInstances()
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return fmt.Errorf("no Claude Code installation found. Run `system-general-ai install --target claude` first")
	}

	syncOpts := claude.SyncOptions{
		Templates:    templates,
		Mode:         mode,
		EnableSDD:    true,
		EngramBinary: "", // leave engram MCP block as-is
	}

	out := cmd.OutOrStdout()
	for _, inst := range instances {
		if err := claude.Sync(inst, syncOpts); err != nil {
			return fmt.Errorf("apply to %s: %w", inst.Path, err)
		}
		fmt.Fprintf(out, "  done: %s - %s\n", inst.Path, mode)
	}
	return nil
}

func runConfigurePersona(cmd *cobra.Command, args []string) error {
	persona := args[0]
	if persona == "-" || persona == "neutral" {
		persona = claude.PersonaNeutral
	}

	out := cmd.OutOrStdout()
	errOut := cmd.OutOrStderr()
	displayed := persona
	if displayed == "" {
		displayed = "(neutral - outputStyle removed)"
	}

	targets, err := normalizeTargets(configurePersonaTargets, configurePersonaAll, allTargets)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	engramPath := filepath.Join(defaultBinDir(), engramBinaryName())
	successes := 0

	for _, target := range targets {
		switch target {
		case "claude":
			instances, err := claude.DetectInstances()
			if err != nil {
				fmt.Fprintf(errOut, "Warning: failed to detect Claude Code: %v\n", err)
				continue
			}
			if len(instances) == 0 {
				fmt.Fprintln(errOut, "Warning: no Claude Code installation found.")
				continue
			}
			if err := claude.ApplyPersonaToInstances(instances, persona); err != nil {
				fmt.Fprintf(errOut, "Warning: failed to apply persona to Claude Code: %v\n", err)
				continue
			}
			successes += len(instances)
			for _, inst := range instances {
				fmt.Fprintf(out, "  done: Claude %s - %s\n", inst.Path, displayed)
			}
		case "gemini":
			if err := gemini.Sync(cwd, gemini.SyncOptions{
				Templates:    templates,
				EnableSDD:    true,
				Persona:      persona,
				EngramBinary: engramPath,
			}); err != nil {
				fmt.Fprintf(errOut, "Warning: failed to sync persona to Gemini CLI: %v\n", err)
				continue
			}
			successes++
			fmt.Fprintf(out, "  done: Gemini - %s\n", displayed)
		case "codex":
			codexInst, err := codexpkg.DetectInstance()
			if err != nil {
				fmt.Fprintf(errOut, "Warning: failed to detect Codex: %v\n", err)
				continue
			}
			if err := codexpkg.Sync(codexInst, codexpkg.SyncOptions{
				Templates:    templates,
				ProjectDir:   cwd,
				EnableSDD:    true,
				Persona:      persona,
				EngramBinary: engramPath,
			}); err != nil {
				fmt.Fprintf(errOut, "Warning: failed to sync persona to Codex: %v\n", err)
				continue
			}
			successes++
			fmt.Fprintf(out, "  done: Codex %s - %s\n", codexInst.Path, displayed)
		}
	}

	if successes == 0 {
		return fmt.Errorf("persona was not applied to any target")
	}
	return nil
}

func isWindows() bool { return runtime.GOOS == "windows" }
