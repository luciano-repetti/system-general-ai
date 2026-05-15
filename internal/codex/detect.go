package codex

import (
	"fmt"
	"os"
	"path/filepath"
)

// Instance is a Codex configuration root.
type Instance struct {
	Path string
}

// ConfigPath returns ~/.codex/config.toml.
func (i Instance) ConfigPath() string {
	return filepath.Join(i.Path, "config.toml")
}

// AGENTSPath returns the global Codex AGENTS.md.
func (i Instance) AGENTSPath() string {
	return filepath.Join(i.Path, "AGENTS.md")
}

// SkillsDir returns the global Codex skills directory.
func (i Instance) SkillsDir() string {
	return filepath.Join(i.Path, "skills")
}

// EngramInstructionsPath returns the Codex model instructions file for Engram.
func (i Instance) EngramInstructionsPath() string {
	return filepath.Join(i.Path, "engram-instructions.md")
}

// EngramCompactPromptPath returns the Codex compact prompt override for Engram.
func (i Instance) EngramCompactPromptPath() string {
	return filepath.Join(i.Path, "engram-compact-prompt.md")
}

// DetectInstance returns the Codex config root.
//
// CODEX_HOME is treated as an override when present; otherwise ~/.codex is used.
// Unlike Claude detection, the directory does not need to already exist: sync can
// create it for a fresh Codex install.
func DetectInstance() (Instance, error) {
	if env := os.Getenv("CODEX_HOME"); env != "" {
		abs, err := filepath.Abs(env)
		if err != nil {
			return Instance{}, fmt.Errorf("resolve CODEX_HOME: %w", err)
		}
		return Instance{Path: abs}, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Instance{}, fmt.Errorf("resolve home dir: %w", err)
	}
	return Instance{Path: filepath.Join(home, ".codex")}, nil
}
