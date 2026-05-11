---
name: sdd-init
description: Initialize SDD context for a project. Detect stack, test runner, conventions. Save under sdd-init/{project} in Engram.
trigger: First time SDD is used in a project, or when stack changed (new test runner, new dependencies).
---

# sdd-init

Bootstrap the SDD context for the current project. Run silently when needed; results are persisted in Engram.

## Steps

1. Detect language and framework:
   - Read package.json, go.mod, requirements.txt, pyproject.toml, Cargo.toml, *.csproj, pom.xml, build.gradle
   - Identify primary language and major framework

2. Detect test runner and build/dev commands:
   - Look at scripts (package.json), targets (Makefile, Taskfile), .github/workflows
   - Capture exact commands: `test_command`, `build_command`, `dev_command`, `install_command`

3. Detect strict TDD eligibility:
   - `strict_tdd: true` only if a test runner exists AND can run a single test file in under ~30s
   - Otherwise `strict_tdd: false`

4. Detect conventions:
   - Look at existing files for naming patterns, folder structure, import style
   - Note any CLAUDE.md or AGENTS.md hints already present

5. Save to Engram with `topic_key: sdd-init/{project}`, scope `project`. Include:
   - language, framework
   - test_command, build_command, dev_command, install_command
   - strict_tdd (true/false)
   - conventions (free-form)
   - detected_files (list of config files inspected)

## Output

Return a one-paragraph executive summary plus the saved Engram observation ID.
