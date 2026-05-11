package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/lucianorepetti/system-general-ai/internal/claude"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Change post-install settings (permissions, persona)",
}

var configurePermissionsCmd = &cobra.Command{
	Use:   "permissions <permissive|balanced|strict>",
	Short: "Switch the permission profile for every detected instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigurePermissions,
}

var configurePersonaCmd = &cobra.Command{
	Use:   "persona <name>",
	Short: "Switch the active persona. Pass an empty string to remove it (use \"-\" or \"neutral\")",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigurePersona,
}

func init() {
	configureCmd.AddCommand(configurePermissionsCmd)
	configureCmd.AddCommand(configurePersonaCmd)
}

func runConfigurePermissions(cmd *cobra.Command, args []string) error {
	mode := claude.PermissionMode(args[0])
	switch mode {
	case claude.ModePermissive, claude.ModeBalanced, claude.ModeStrict:
		// ok
	default:
		return fmt.Errorf("unknown mode %q — expected one of: permissive, balanced, strict", args[0])
	}

	instances, err := claude.DetectInstances()
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return fmt.Errorf("no Claude Code installation found. Run `system-general-ai install` first")
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
		fmt.Fprintf(out, "  ✓ %s — %s\n", inst.Path, mode)
	}
	return nil
}

func runConfigurePersona(cmd *cobra.Command, args []string) error {
	persona := args[0]
	if persona == "-" || persona == "neutral" {
		persona = claude.PersonaNeutral
	}

	instances, err := claude.DetectInstances()
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return fmt.Errorf("no Claude Code installation found. Run `system-general-ai install` first")
	}

	if err := claude.ApplyPersonaToInstances(instances, persona); err != nil {
		return fmt.Errorf("apply persona: %w", err)
	}

	out := cmd.OutOrStdout()
	displayed := persona
	if displayed == "" {
		displayed = "(neutral — outputStyle removed)"
	}
	for _, inst := range instances {
		fmt.Fprintf(out, "  ✓ %s — %s\n", inst.Path, displayed)
	}
	return nil
}

func isWindows() bool { return runtime.GOOS == "windows" }

