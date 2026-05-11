package main

import (
	"fmt"
	"io/fs"
	"os"

	sgai "github.com/lucianorepetti/system-general-ai"
	"github.com/lucianorepetti/system-general-ai/internal/cli"
)

func main() {
	// Strip the "templates" prefix so consumers (claude.Sync) see a clean root.
	templates, err := fs.Sub(sgai.Templates, "templates")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to mount embedded templates:", err)
		os.Exit(1)
	}

	if err := cli.Execute(templates); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

