---
parent: specs/2026-10-15-codex-harness-design
closes: [DBT-0122]
depth: minimal
id: PLN-0151
created: "2026-10-15 16:35:17"
hash: aob4ljd
started: "2026-10-15 16:37:32"
finished: "2026-10-15 16:54:02"
---
# Run acta inside Codex Implementation Plan

**Goal:** A user who opens Codex in an acta repo gets the acta skills, session note, output style, wiki hints and second-brainstorm block, with Claude Code behaviour unchanged.

**Spec:** .acta/specs/2026-10-15-codex-harness-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- No signal of Codex means Claude Code: every hook output under Claude Code input stays byte for byte what it is today, except the moved subagent rules (Task 3).
- Target is codex-cli 0.162.1. Facts measured in Codex live in `plugin/codex/FACTS.md`.
- Codex signal: env `PLUGIN_ROOT` is set (FACTS.md row H).
- Plugin files must stay embedded in the binary (`plugin/embed.go`).

## Ruling 2026-10-15 after Task 1

The spike showed Codex reads `.claude-plugin/marketplace.json` and runs the plugin hooks as they are. The user ruled: no Codex `plugin.json`, no fourth version file, no `acta setup` fallback. The old Task 2 (manifest version) is dropped; later tasks are renumbered.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2, Task 4
- Wave 3: Task 3
- Wave 4: Task 5

### Task 1: Spike, Codex plugin hooks and harness signal

**Files:** Create `plugin/codex/FACTS.md`

**Seam:** `codex plugin marketplace add <plugin dir>` then `codex exec` in a scratch repo

**verify:** FACTS.md answers each open question with the command run and its output: do plugin hooks fire for SessionStart, UserPromptSubmit, PreToolUse and PostToolUse; the exact signal present only under Codex; the hook input JSON for Bash and apply_patch.

- [x] Failing test: none, spike; the open questions are the red state.
- [x] Change: install in a throwaway `CODEX_HOME`, log hook stdin and env, record results.
- [x] Commit: `docs(codex): record measured Codex plugin facts`.

### Task 2: Hook reads Codex input

**Files:** Modify `internal/hook/wikihint.go`, `internal/hook/hook.go`, `plugin/hooks/hooks.json`; Test `internal/hook/wikihint_test.go`, `internal/hook/hook_test.go`

**Seam:** `WikiHints(root, repo, ToolEvent)`, `hintFiles`, a new `IsCodex() bool`

**verify:** Every file an apply_patch touches (Update, Add, Delete, Move to) gets its wiki hint, and no other text in the patch is ever read as a path. Adversarial inputs: a content line that quotes a header, CRLF line ends, paths with spaces, `..` paths outside the repo. `IsCodex` is true only when `PLUGIN_ROOT` is set. The PreToolUse matcher in hooks.json includes `apply_patch`.

- [x] Failing test: `hintFiles` on an apply_patch event returns no paths today.
- [x] Change: parse patch headers in `hintFiles`, add `IsCodex`, add `apply_patch` to the matcher.
- [x] Commit: `feat(hook): wiki hints for Codex apply_patch edits`.

### Task 3: Session note carries style and subagent rules; skills go neutral

**Files:** Modify `internal/hook/hook.go`, `plugin/skills/{slice,shape,review,build,debug}/SKILL.md`; Test `internal/hook/hook_test.go`, create `internal/plugincheck/neutral_skills_test.go`; update the skill tests in `internal/plugincheck/skill_*_test.go` that pin the old wording

**Seam:** `SessionStart(Input) string`

**verify:** Under Codex the note holds the output style text from `plugin/output-styles/` and the Codex subagent rule; under Claude Code it holds the split model rule and no style text. No skill under `plugin/skills/**` names `Agent` tool, `subagent_type`, `model: "sonnet"` or `model: "opus"` in any form, and each rule the skills lost is now in the note. List each moved rule.

- [x] Failing test: SessionStart under Codex has no style text, and the neutral-skills guard finds Claude-only names.
- [x] Change: add the style and subagent blocks to SessionStart; reword the five skills.
- [x] Commit: `feat(hook): Codex session note carries style and subagent rules`.

### Task 4: Doctor checks Codex

**Files:** Modify `internal/doctor/doctor.go`; Test `internal/doctor/doctor_test.go`

**Seam:** the doctor check list behind `acta doctor`

**verify:** With `codex` not on PATH doctor says nothing about Codex. With `codex` on PATH, doctor reports whether the acta plugin is installed and tells the user to trust its hooks in `/hooks`; it never fails the run or changes files outside the repo. List each state checked.

- [x] Failing test: doctor with a fake `codex` on PATH reports nothing about Codex today.
- [x] Change: add a Codex check that reads `codex plugin list`.
- [x] Commit: `feat(doctor): check the acta plugin in Codex`.

### Task 5: Docs and wiki

**Files:** Modify `plugin/README.md`, `plugin/codex/FACTS.md`; Create `.acta/wiki/codex-harness.md`

**Seam:** `go test ./internal/plugincheck` (doc guards) and `acta wiki ls`

**verify:** A fresh user can install acta in Codex from README alone, including the install from GitHub (tested: with or without `--sparse plugin`) and the `/hooks` trust step; the wiki page names how acta detects Codex and each way its hooks differ from Claude Code, and its paths cover `internal/hook` and `plugin/hooks`. The manual smoke run (note, wiki hint on edit, second brainstorm blocked) is recorded in FACTS.md with output.

- [x] Failing test: none, docs; plugincheck guards must stay green.
- [x] Change: README Codex part, wiki page, smoke run in FACTS.md.
- [x] Commit: `docs(codex): install guide and codex-harness wiki page`.

## Rulings during build

- Task 4: the doctor check reads Codex's `config.toml` (the `[plugins."acta@..."]` table) instead of running `codex plugin list`, so doctor stays fast and offline. `checkHarness` passes when acta is in Codex only, so a Codex-only user does not get a false fail.
- Task 5: `codex plugin marketplace add iyay/acta` fails (FACTS.md, "Install from GitHub"), so `acta setup` installs Codex from the unpacked local plugin, like Claude Code. Files added to the task: `internal/setup/plan.go`, `internal/cli/setup_cmd.go` and their tests.

## Polish

### Task 6: Review polish

**verify:** the NOTEs below applied, and nothing else changes.

- [x] Move `codexDir` below `homeDir` so the `homeDir` comment sits on its own function.
- [x] Doctor reads a hand-edited Codex config the same as Codex's own: inline comments, tabs, a quoted `enabled` key, a single-quoted header with spaces.
- [x] Record the Task 4 and Task 5 rulings in this plan.
- [x] Commit: `polish: review notes PLN-0151`

Polish reverted (afd8a90): the polish review found a BLOCKER, the config parser no longer stripped \r, so CRLF config files read as not installed. The parser item moved to debt; the codexDir comment move is a low note and was dropped with the revert.

## Review notes

Round 1 CLEAN on both axes. Notes kept as notes: IsCodex trusts any PLUGIN_ROOT; patchFiles leaves paths outside the repo to the wiki match, which finds no page; checkHarness stops at Codex before the omp link check; codexDir sits between homeDir and its comment.

## Fix round 1

User review found two P2s after round 1 was CLEAN.

### Task 7: Fix round 1

**Files:** Modify `internal/cli/hook.go`, `internal/doctor/codex.go`; Test `internal/cli/hook_test.go`, `internal/doctor/codex_test.go`

**verify:** No apply_patch call is ever blocked for its file content: the brainstorm and test-runner guards read shell commands only, and a patch still gets its wiki hint. Doctor reads every form of `enabled = true` a user may write in the acta table (trailing comment, tab, CRLF, quoted key, single-quoted header) as installed, and a commented-out line or a non-true value as not. Closes DBT-0122.

- [x] Failing test: patch content lines that look like blocked commands exit 2; hand-written TOML forms read as not installed.
- [x] Change: skip both shell guards for apply_patch; read config lines with comment, whitespace and quote handling.
- [x] Commit: `fix(hook): patch content never trips shell guards; doctor reads hand-written TOML`.

## State

### Next

Land paused 2026-10-15: user reviews the worktree first. All gates green except land itself: full suite exit 0, wiki check exit 0, evals 19 cases (answers-appended flaked 1 of 9 runs on branch, 5/5 on last rerun, 3/3 on main). Next: after user review, run acta:land in tidy mode.
