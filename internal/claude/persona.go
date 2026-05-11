package claude

import (
	"fmt"
)

// Persona names used by system-general-ai.
//
// The default Claude Code "neutral" persona has NO outputStyle key set —
// not an empty string, not a null. A previously observed bug in another
// wrapper set the field to "" instead of removing the key, leaving stale
// behavior in place. We avoid that pitfall here by deleting the key.
const (
	PersonaSystemClaudeAI = "system-general-ai"
	PersonaNeutral        = "" // empty string means: REMOVE the outputStyle key
)

// ApplyPersona writes the desired persona into a Claude Code settings.json.
//
// Behavior:
//   - persona == PersonaNeutral (""): the "outputStyle" key is REMOVED from settings.json.
//     This is the fix for upstream issue #204.
//   - persona is any non-empty string: settings.outputStyle is set to that value.
//
// User-authored keys in settings.json are preserved.
func ApplyPersona(settingsPath, persona string) error {
	existing, err := readJSONOrEmpty(settingsPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", settingsPath, err)
	}

	if persona == PersonaNeutral {
		delete(existing, "outputStyle")
	} else {
		existing["outputStyle"] = persona
	}

	if err := writeJSON(settingsPath, existing); err != nil {
		return fmt.Errorf("write %s: %w", settingsPath, err)
	}
	return nil
}

// ApplyPersonaToInstances applies the given persona to every instance.
// Returns the first error encountered, but tries every instance.
func ApplyPersonaToInstances(instances []Instance, persona string) error {
	var firstErr error
	for _, inst := range instances {
		if err := ApplyPersona(inst.SettingsPath(), persona); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("instance %s: %w", inst.Path, err)
		}
	}
	return firstErr
}

