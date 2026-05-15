package cli

import (
	"fmt"
	"strings"
)

var allTargets = []string{"claude", "gemini", "codex"}

func normalizeTargets(requested []string, all bool, defaults []string) ([]string, error) {
	if all {
		return append([]string(nil), allTargets...), nil
	}
	if len(requested) == 0 {
		if len(defaults) == 0 {
			return nil, fmt.Errorf("no target platforms selected")
		}
		return append([]string(nil), defaults...), nil
	}

	seen := map[string]bool{}
	out := make([]string, 0, len(requested))
	for _, raw := range requested {
		for _, part := range strings.Split(raw, ",") {
			target := strings.ToLower(strings.TrimSpace(part))
			if target == "" {
				continue
			}
			if target == "all" {
				return append([]string(nil), allTargets...), nil
			}
			if !isKnownTarget(target) {
				return nil, fmt.Errorf("unknown target %q; expected claude, gemini, codex, or all", target)
			}
			if !seen[target] {
				seen[target] = true
				out = append(out, target)
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no target platforms selected")
	}
	return out, nil
}

func isKnownTarget(target string) bool {
	for _, known := range allTargets {
		if target == known {
			return true
		}
	}
	return false
}
