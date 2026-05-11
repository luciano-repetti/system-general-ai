// Package sgai is the module root. It exists solely to host the embed.FS
// that ships the templates/ folder inside the binary.
package sgai

import "embed"

// Templates is the embedded template tree, rooted at the templates/ directory.
// Other packages (notably internal/claude.Sync) consume it as an fs.FS.
//
//go:embed all:templates
var Templates embed.FS

