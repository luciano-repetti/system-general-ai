package cli

import (
	"io/fs"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags="-X .../internal/cli.version=v0.1.0".
var version = "dev"

// templates is set by Execute and consumed by subcommands.
var templates fs.FS

var rootCmd = &cobra.Command{
	Use:           "system-general-ai",
	Short:         "Direct, focused AI ecosystem configurator for Claude Code",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(uninstallCmd)
}

// Execute is the entrypoint called by cmd/system-general-ai/main.go.
// The fs.FS holds the embedded templates tree.
func Execute(t fs.FS) error {
	templates = t
	return rootCmd.Execute()
}

