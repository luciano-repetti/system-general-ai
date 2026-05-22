package codex

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	markerStartFmt = "<!-- BEGIN: system-general-ai/%s -->"
	markerEndFmt   = "<!-- END: system-general-ai/%s -->"
)

func mergeMarkedSection(existing, sectionName, content string) string {
	startMarker := fmt.Sprintf(markerStartFmt, sectionName)
	endMarker := fmt.Sprintf(markerEndFmt, sectionName)
	newSection := fmt.Sprintf("%s\n%s\n%s", startMarker, strings.TrimSpace(content), endMarker)

	if strings.TrimSpace(existing) == "" {
		return newSection + "\n"
	}

	startIdx := strings.Index(existing, startMarker)
	endIdx := strings.Index(existing, endMarker)
	if startIdx >= 0 && endIdx > startIdx {
		before := existing[:startIdx]
		after := existing[endIdx+len(endMarker):]
		return before + newSection + after
	}

	sep := "\n\n"
	if strings.HasSuffix(existing, "\n\n") {
		sep = ""
	} else if strings.HasSuffix(existing, "\n") {
		sep = "\n"
	}
	return existing + sep + newSection + "\n"
}

func removeMarkedSection(existing, sectionName string) string {
	startMarker := fmt.Sprintf(markerStartFmt, sectionName)
	endMarker := fmt.Sprintf(markerEndFmt, sectionName)
	startIdx := strings.Index(existing, startMarker)
	endIdx := strings.Index(existing, endMarker)
	if startIdx < 0 || endIdx <= startIdx {
		return existing
	}

	before := strings.TrimRight(existing[:startIdx], "\n")
	after := strings.TrimLeft(existing[endIdx+len(endMarker):], "\n")
	if before == "" {
		if after == "" {
			return ""
		}
		return after
	}
	if after == "" {
		return before + "\n"
	}
	return before + "\n\n" + after
}

func upsertCodexEngramBlock(content, engramCmd string) string {
	if strings.TrimSpace(engramCmd) == "" {
		engramCmd = "engram"
	}
	escapedCmd := tomlQuote(engramCmd)
	block := "[mcp_servers.engram]\ncommand = " + escapedCmd + "\nargs = [\"mcp\", \"--tools=agent\"]\nenabled = true\nstartup_timeout_sec = 20\ntool_timeout_sec = 60"

	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	kept := make([]string, 0, len(lines))

	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "[mcp_servers.engram]" {
			i++
			for i < len(lines) {
				next := strings.TrimSpace(lines[i])
				if strings.HasPrefix(next, "[") && strings.HasSuffix(next, "]") {
					break
				}
				i++
			}
			continue
		}
		kept = append(kept, lines[i])
		i++
	}

	base := strings.TrimSpace(strings.Join(kept, "\n"))
	if base == "" {
		return block + "\n"
	}
	return base + "\n\n" + block + "\n"
}

func removeCodexEngramBlock(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	kept := make([]string, 0, len(lines))
	for i := 0; i < len(lines); {
		if strings.TrimSpace(lines[i]) == "[mcp_servers.engram]" {
			i++
			for i < len(lines) {
				next := strings.TrimSpace(lines[i])
				if strings.HasPrefix(next, "[") && strings.HasSuffix(next, "]") {
					break
				}
				i++
			}
			continue
		}
		kept = append(kept, lines[i])
		i++
	}
	out := strings.TrimSpace(strings.Join(kept, "\n"))
	if out == "" {
		return ""
	}
	return out + "\n"
}

func upsertTopLevelTOMLString(content, key, value string) string {
	return upsertTopLevelTOMLStrings(content, []tomlStringKV{{key: key, value: value}})
}

type tomlStringKV struct {
	key   string
	value string
}

func upsertTopLevelTOMLStrings(content string, pairs []tomlStringKV) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	managed := make(map[string]struct{}, len(pairs))
	for _, pair := range pairs {
		managed[pair.key] = struct{}{}
	}

	cleaned := make([]string, 0, len(lines)+len(pairs))
	inTopLevel := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isTOMLTableHeader(trimmed) {
			inTopLevel = false
		}
		if inTopLevel {
			skip := false
			for key := range managed {
				if isTOMLKeyLine(trimmed, key) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		cleaned = append(cleaned, line)
	}

	for len(cleaned) > 0 && strings.TrimSpace(cleaned[0]) == "" {
		cleaned = cleaned[1:]
	}
	for len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) == "" {
		cleaned = cleaned[:len(cleaned)-1]
	}

	insertAt := len(cleaned)
	for i, line := range cleaned {
		trimmed := strings.TrimSpace(line)
		if isTOMLTableHeader(trimmed) {
			insertAt = i
			break
		}
	}

	newLines := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		if strings.TrimSpace(pair.value) == "" {
			continue
		}
		newLines = append(newLines, fmt.Sprintf("%s = %s", pair.key, tomlQuote(pair.value)))
	}

	if insertAt < len(cleaned) && len(newLines) > 0 {
		if insertAt > 0 && strings.TrimSpace(cleaned[insertAt-1]) != "" {
			newLines = append(newLines, "")
		}
	} else if insertAt == len(cleaned) && len(cleaned) > 0 && len(newLines) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) != "" {
		newLines = append([]string{""}, newLines...)
	}

	out := make([]string, 0, len(cleaned)+1)
	out = append(out, cleaned[:insertAt]...)
	out = append(out, newLines...)
	out = append(out, cleaned[insertAt:]...)
	return strings.TrimSpace(strings.Join(out, "\n")) + "\n"
}

func removeTopLevelTOMLKey(content, key string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	inTopLevel := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isTOMLTableHeader(trimmed) {
			inTopLevel = false
		}
		if inTopLevel && isTOMLKeyLine(trimmed, key) {
			continue
		}
		out = append(out, line)
	}
	result := strings.TrimSpace(strings.Join(out, "\n"))
	if result == "" {
		return ""
	}
	return result + "\n"
}

func tomlQuote(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + replacer.Replace(s) + `"`
}

func validateTOML(content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	var decoded map[string]any
	if _, err := toml.Decode(content, &decoded); err != nil {
		return err
	}
	return nil
}

func isTOMLTableHeader(trimmed string) bool {
	if strings.HasPrefix(trimmed, "#") {
		return false
	}
	return strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
}

func isTOMLKeyLine(trimmed, key string) bool {
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}
	return strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=")
}
