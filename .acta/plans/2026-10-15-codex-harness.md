---
parent: specs/2026-10-15-codex-harness-design
depth: minimal
id: PLN-0151
created: "2026-10-15 16:35:17"
hash: aob4ljd
---
# Run acta inside Codex Implementation Plan

**Goal:** A user who opens Codex in an acta repo gets the acta skills, session note, output style, wiki hints and second-brainstorm block, with Claude Code behaviour unchanged.

**Spec:** .acta/specs/2026-10-15-codex-harness-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- No signal of Codex means Claude Code: every hook output under Claude Code input stays byte for byte what it is today, except the moved subagent rules (Task 4).
- Target is codex-cli 0.162.1. Facts measured in Codex go to `plugin/codex/FACTS.md`.
- Plugin files must stay embedded in the binary (`plugin/embed.go`).

## Waves

- Wave 1: Task 1
- Wave 2: Task 2, Task 3
- Wave 3: Task 4, Task 5
- Wave 4: Task 6

### Task 1: Spike, Codex plugin hooks and harness signal

**Files:** Create `plugin/codex/FACTS.md`, `plugin/.codex-plugin/plugin.json`, `.agents/plugins/marketplace.json`

**Seam:** `codex plugin marketplace add <repo path>` then a Codex session in a scratch repo

**verify:** FACTS.md answers each open question with the command run and its output: do plugin hooks fire for SessionStart, UserPromptSubmit, PreToolUse and PostToolUse; the exact env var or input field present only under Codex; the hook input JSON for Bash and apply_patch. No question is answered from docs alone.

Install the plugin from this repo into Codex, log each hook's stdin and env to a scratch file, and record the results. If plugin hooks do not fire, say so in FACTS.md and stop the plan for a ruling before Task 5's fallback.

### Task 2: Codex manifest version kept in step

**Files:** Modify `scripts/release`, `internal/plugincheck/release_test.go`, `internal/plugincheck/plugin_test.go`, `plugin/embed.go` (only if the new files are not embedded)

**Seam:** `go test ./internal/plugincheck -run 'TestRelease|TestPlugin'`

**verify:** No release can leave the four version files apart: plugincheck fails when the Codex `plugin.json` is missing, does not parse, or names a version that differs from the other three, and `scripts/release` bumps all four. List each failure input tried.

### Task 3: Hook reads Codex input

**Files:** Modify `internal/hook/wikihint.go`, `internal/hook/hook.go`, `plugin/hooks/hooks.json`; Test `internal/hook/wikihint_test.go`, `internal/hook/hook_test.go`

**Seam:** `WikiHints(root, repo, ToolEvent)`, `hintFiles`, a new `IsCodex(Input) bool`

**verify:** Every file an apply_patch touches (Update, Add, Delete, Move to) gets its wiki hint, and no other text in the patch is ever read as a path. Adversarial inputs: a patch line that quotes a header inside file content, CRLF line ends, paths with spaces, `..` paths outside the repo. `IsCodex` is true only on the signal Task 1 found. The PreToolUse matcher in hooks.json includes `apply_patch`.

### Task 4: Session note carries style and subagent rules; skills go neutral

**Files:** Modify `internal/hook/hook.go`, `plugin/skills/{slice,shape,review,build,debug}/SKILL.md`; Test `internal/hook/hook_test.go`, create `internal/plugincheck/neutral_skills_test.go`; update the skill tests in `internal/plugincheck/skill_*_test.go` that pin the old wording

**Seam:** `SessionStart(Input) string`

**verify:** Under Codex the note holds the output style text from `plugin/output-styles/` and the Codex subagent rule; under Claude Code it holds the split model rule and no style text. No skill under `plugin/skills/**` names `Agent` tool, `subagent_type`, `model: "sonnet"` or `model: "opus"` in any form, and each rule the skills lost is now in the note. List each moved rule.

### Task 5: Doctor checks Codex

**Files:** Modify `internal/doctor/doctor.go`; Test `internal/doctor/doctor_test.go`

**Seam:** the doctor check list behind `acta doctor`

**verify:** With `codex` not on PATH doctor says nothing about Codex. With `codex` on PATH, doctor reports whether the acta plugin is installed, and never fails or changes files outside the repo. List each state checked. If Task 1 found plugin hooks do not fire, this task also adds the `acta setup` fallback, after a ruling.

### Task 6: Docs and wiki

**Files:** Modify `plugin/README.md`, `plugin/codex/FACTS.md`; Create `.acta/wiki/codex-harness.md`

**Seam:** `go test ./internal/plugincheck` (doc guards) and `acta wiki ls`

**verify:** A fresh user can install acta in Codex from README alone; the wiki page names how acta detects Codex and each way its hooks differ from Claude Code, and its paths cover `internal/hook` and `plugin/hooks`. The manual smoke run (note, wiki hint on edit, second brainstorm blocked) is recorded in FACTS.md with output.
