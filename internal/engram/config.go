package engram

import (
	"encoding/json"
	"fmt"
)

// MCPConfig is the structure expected by Claude Code's settings.json under "mcpServers.engram".
type MCPConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// BuildMCPConfig produces the MCP configuration block for engram given its absolute path.
func BuildMCPConfig(binaryPath string) MCPConfig {
	return MCPConfig{
		Command: binaryPath,
		Args:    []string{"serve"},
		Env:     map[string]string{},
	}
}

// MarshalSnippet returns the engram MCP server snippet ready to merge into settings.json.
// Output shape:
//
//	{ "mcpServers": { "engram": { ... } } }
func MarshalSnippet(cfg MCPConfig) ([]byte, error) {
	wrapper := map[string]any{
		"mcpServers": map[string]any{
			"engram": cfg,
		},
	}
	out, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal snippet: %w", err)
	}
	return out, nil
}
