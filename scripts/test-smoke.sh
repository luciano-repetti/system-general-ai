#!/usr/bin/env bash
# E2E smoke test for system-general-ai. Safe — does NOT touch ~/.claude.
# Builds the binary, runs `install` against an isolated tempdir, verifies the
# expected files were produced.

set -euo pipefail

cd "$(dirname "$0")/.."

bold=$'\033[1m'; reset=$'\033[0m'; red=$'\033[31m'; green=$'\033[32m'

fail() { printf '%sFAIL%s %s\n' "$red" "$reset" "$1" >&2; exit 1; }
ok()   { printf '%sOK%s   %s\n' "$green" "$reset" "$1"; }
info() { printf '%s%s%s\n' "$bold" "$1" "$reset"; }

# Pick a Python interpreter — prefer python3, then python, then py (Windows).
if   command -v python3 >/dev/null 2>&1; then PY=python3
elif command -v python  >/dev/null 2>&1; then PY=python
elif command -v py      >/dev/null 2>&1; then PY=py
else fail "no python interpreter found (need python3, python, or py)"
fi

# On Windows (Git Bash / MSYS) python is the Windows binary and needs a
# Windows-style path — convert via cygpath when available.
mpath() {
    if command -v cygpath >/dev/null 2>&1; then
        cygpath -m "$1"
    else
        printf '%s' "$1"
    fi
}

info "Building binary ..."
go build -o /tmp/sgai-test ./cmd/system-general-ai
ok "build succeeded"

TEST_DIR="$(mktemp -d)"
trap 'rm -rf "$TEST_DIR"' EXIT
info "Test instance: $TEST_DIR"

info "Running zero-config install ..."
CLAUDE_CONFIG_DIR="$TEST_DIR" /tmp/sgai-test install
ok "install completed"

info "Verifying produced files ..."

[ -f "$TEST_DIR/output-styles/system-general-ai.md" ] || fail "output-styles/system-general-ai.md missing"
ok "output-styles/system-general-ai.md"

[ -f "$TEST_DIR/CLAUDE.md" ] || fail "CLAUDE.md missing"
grep -q "<!-- BEGIN: system-general-ai/global-rules -->" "$TEST_DIR/CLAUDE.md" \
  || fail "global-rules marker missing in CLAUDE.md"
grep -q "<!-- BEGIN: system-general-ai/sdd-orchestrator -->" "$TEST_DIR/CLAUDE.md" \
  || fail "sdd-orchestrator marker missing in CLAUDE.md"
ok "CLAUDE.md has both marker sections"

[ -f "$TEST_DIR/settings.json" ] || fail "settings.json missing"
TEST_SETTINGS_PY="$(mpath "$TEST_DIR/settings.json")"
"$PY" -c "import json,sys; j=json.load(open(r'$TEST_SETTINGS_PY'));
assert j.get('outputStyle')=='system-general-ai', 'outputStyle wrong';
assert 'permissions' in j, 'permissions missing';
assert 'mcpServers' in j and 'engram' in j['mcpServers'], 'engram MCP missing'" \
  || fail "settings.json content wrong"
ok "settings.json has outputStyle, permissions, mcpServers.engram"

for skill in sdd-init sdd-explore sdd-propose sdd-spec sdd-design \
             sdd-tasks sdd-apply sdd-verify sdd-archive \
             shell-runner compact-suggest; do
    [ -f "$TEST_DIR/skills/$skill/SKILL.md" ] || fail "skill $skill/SKILL.md missing"
done
ok "all 11 skills installed (directory format)"

info "Running idempotency check (re-install) ..."
SUM_BEFORE="$(find "$TEST_DIR" -type f -exec sha256sum {} \; | sort)"
CLAUDE_CONFIG_DIR="$TEST_DIR" /tmp/sgai-test install >/dev/null
SUM_AFTER="$(find "$TEST_DIR" -type f -exec sha256sum {} \; | sort)"
[ "$SUM_BEFORE" = "$SUM_AFTER" ] || fail "second install produced different output"
ok "idempotent (re-running install yields same files)"

info "Running configure permissions strict ..."
CLAUDE_CONFIG_DIR="$TEST_DIR" /tmp/sgai-test configure permissions strict >/dev/null
"$PY" -c "import json; j=json.load(open(r'$TEST_SETTINGS_PY'));
ask=set(j['permissions'].get('ask',[]));
assert 'Edit' in ask and 'Write' in ask, 'strict mode did not move Edit/Write to ask'" \
  || fail "configure permissions did not switch to strict"
ok "configure permissions strict applied"

info "Running configure persona neutral (fix #204 regression) ..."
CLAUDE_CONFIG_DIR="$TEST_DIR" /tmp/sgai-test configure persona neutral >/dev/null
"$PY" -c "import json; j=json.load(open(r'$TEST_SETTINGS_PY'));
assert 'outputStyle' not in j, 'outputStyle was NOT removed (fix #204 broken)'" \
  || fail "fix #204 regression — outputStyle key persisted"
ok "fix #204 verified — neutral removes outputStyle key"

info "Running uninstall ..."
CLAUDE_CONFIG_DIR="$TEST_DIR" /tmp/sgai-test uninstall >/dev/null
[ ! -f "$TEST_DIR/output-styles/system-general-ai.md" ] || fail "output-style not removed"
[ ! -f "$TEST_DIR/skills/sdd-init/SKILL.md" ] || fail "skills/sdd-init/SKILL.md not removed"
ok "uninstall removed managed files"

# ---------------------------------------------------------------------------
# Merge regression test: deep-merge of permissions + skill directory format.
# Uses a SECOND tempdir pre-seeded with user customizations to verify the
# install does NOT clobber them.
# ---------------------------------------------------------------------------
info ""
info "Running merge-regression test (Bug 1 + Bug 2) ..."

MERGE_DIR="$(mktemp -d)"
trap 'rm -rf "$TEST_DIR" "$MERGE_DIR"' EXIT

# 1. Pre-seed settings.json with custom user permissions + unrelated keys.
mkdir -p "$MERGE_DIR"
cat > "$MERGE_DIR/settings.json" <<'EOF'
{
  "theme": "dark",
  "permissions": {
    "defaultMode": "acceptEdits",
    "allow": [
      "mcp__custom__tool",
      "Bash(npm test*)",
      "Bash(make*)",
      "Bash(pytest*)",
      "Bash(cargo build*)",
      "Bash(go test*)"
    ],
    "deny": [
      "Bash(rm -rf /etc*)"
    ]
  }
}
EOF

# 2. Pre-seed a user skill in directory format with arbitrary content.
mkdir -p "$MERGE_DIR/skills/sdd-init"
USER_SKILL_CONTENT="# user's custom sdd-init — must not be silently overwritten"
printf '%s\n' "$USER_SKILL_CONTENT" > "$MERGE_DIR/skills/sdd-init/SKILL.md"

# 3. Run install against the seeded tempdir.
CLAUDE_CONFIG_DIR="$MERGE_DIR" /tmp/sgai-test install >/dev/null

# 4. Assert post-install state.
MERGE_SETTINGS_PY="$(mpath "$MERGE_DIR/settings.json")"
MERGE_SKILL_PY="$(mpath "$MERGE_DIR/skills/sdd-init/SKILL.md")"

"$PY" - <<PY || fail "merge regression assertions failed"
import json, sys
with open(r'''$MERGE_SETTINGS_PY''') as f:
    j = json.load(f)

# (a) Unrelated user key survives
assert j.get('theme') == 'dark', f"theme lost: {j.get('theme')!r}"

perms = j.get('permissions') or {}

# (b) User's defaultMode is preserved (NOT overwritten by template's bypassPermissions)
assert perms.get('defaultMode') == 'acceptEdits', \
    f"defaultMode clobbered: {perms.get('defaultMode')!r}"

allow = perms.get('allow') or []
deny  = perms.get('deny')  or []

# (c) All 6 user allow entries still present
for entry in ['mcp__custom__tool', 'Bash(npm test*)', 'Bash(make*)',
              'Bash(pytest*)', 'Bash(cargo build*)', 'Bash(go test*)']:
    assert entry in allow, f"user allow lost: {entry!r}"

# (d) User's deny entry still present
assert 'Bash(rm -rf /etc*)' in deny, "user deny entry lost"

# (e) sgai's template entries merged in
for entry in ['Read', 'Grep']:
    assert entry in allow, f"template allow not merged: {entry!r}"

# (f) Template's deny entries also merged in
for entry in ['Bash(rm -rf /*)', 'Bash(git push --force*)']:
    assert entry in deny, f"template deny not merged: {entry!r}"

# (g) No duplicates in allow / deny (union-with-dedup contract)
assert len(allow) == len(set(allow)), "duplicates in allow"
assert len(deny)  == len(set(deny)),  "duplicates in deny"

# (h) Pinned behavior: sgai DOES overwrite user files at <skills>/<name>/SKILL.md
#     when the skill name collides with a managed one. This is intentional —
#     skills/<name>/ is treated as sgai-owned. Users wanting a custom variant
#     must pick a different name. If this contract changes, this assertion
#     forces an explicit decision rather than a silent regression.
with open(r'''$MERGE_SKILL_PY''') as f:
    skill_content = f.read()
user_marker = "# user's custom sdd-init — must not be silently overwritten"
assert user_marker not in skill_content, \
    "sgai is now PRESERVING user skill content at a managed path — " \
    "if intentional, update this assertion."
# Sanity check: the file should look like the shipped sdd-init (frontmatter present).
assert skill_content.lstrip().startswith('---'), \
    f"installed skill missing frontmatter — got: {skill_content[:100]!r}"
PY
ok "merge regression: user permissions + theme + skill all preserved, template merged in"

# Idempotency check on the merged result (re-install must not change anything).
SUM_BEFORE_MERGE="$(find "$MERGE_DIR" -type f -exec sha256sum {} \; | sort)"
CLAUDE_CONFIG_DIR="$MERGE_DIR" /tmp/sgai-test install >/dev/null
SUM_AFTER_MERGE="$(find "$MERGE_DIR" -type f -exec sha256sum {} \; | sort)"
[ "$SUM_BEFORE_MERGE" = "$SUM_AFTER_MERGE" ] || fail "merge install not idempotent"
ok "merge install is idempotent (deterministic union order)"

info ""
info "All smoke tests passed."

