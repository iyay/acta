---
parent: specs/2026-10-06-skill-audit-followups
depth: minimal
closes: [DBT-0084.01, DBT-0084.02]
id: PLN-0100
created: "2026-10-06 07:10:03"
hash: bu9zdov
---
# skill audit follow-ups Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** No skill text asks a subagent to chat, quotes a fixed English report line, or sends a debug probe into the real checkout.

**Spec:** `.acta/specs/2026-10-06-skill-audit-followups.md`

**Tests:** `scripts/test ./internal/plugincheck/`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Byte caps in `internal/plugincheck/budget_test.go` stay as they are; shorten the new text, never raise a cap.
- `plugin/skills/debug/defense-in-depth.md` does not change.

## Waves

- Wave 1: Tasks 1 and 2.
- Wave 2: Task 3.

### Task 1: build texts report, not chat

**Files:** `plugin/skills/build/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, `internal/plugincheck/skill_build_test.go`

**verify:** No line in the build skill or implementer prompt asks the subagent to hold a conversation, and no line gives a fixed English sentence to say to the user; list every "ask" and every quoted chat line in both files.

- [ ] Add `MustNot` `"Ask them now"` and `"Already in isolated workspace"` to `TestSkillBuild`; it fails on today's text.
- [ ] In `implementer-prompt.md`, "**Ask them now.** Raise any concerns before starting work." becomes "Stop and report NEEDS_CONTEXT with the question before you start." and the "**While you work:**" lines become "If something is unexpected or unclear, stop and report NEEDS_CONTEXT with the question. Don't guess."; in `SKILL.md`, the two quoted report lines become one line: "Tell the user, in their chat language, the worktree path and branch, or that HEAD is detached and a branch is needed at finish time."
- [ ] Commit: `fix(skills): build asks for NEEDS_CONTEXT and reports in the chat language`

### Task 2: debug probes stay out of the checkout

**Files:** `plugin/skills/debug/root-cause-tracing.md`, `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`

**verify:** No debug text tells the agent to add instrumentation or run a probe in the real checkout, and a CI-only probe always ends with the user running it, never a push; list every instrumentation and probe step in the debug folder and where it runs.

- [ ] Add `MustNot` `"When you can't trace manually, add instrumentation:"` and `Must` `"never push"` to `TestSkillDebug` (confirm the rule reads the folder's reference files); it fails.
- [ ] In `root-cause-tracing.md` the line becomes "When you can't trace manually, instrument a throwaway clone outside the repo:"; in `SKILL.md`, after the CI example's "This reveals" line, add "A probe that can only run in CI: write its steps and ask the user to run them; never push."
- [ ] Commit: `fix(skills): debug probes never touch the checkout or push`

### Task 3: version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same `x.y.z` version, one patch above main's at branch time.

- [ ] Run `scripts/test ./internal/plugincheck/`; the version check already guards agreement, so the bump has no red test of its own.
- [ ] Add 1 to the patch version in all three files.
- [ ] Commit: `chore(plugin): bump patch version`
