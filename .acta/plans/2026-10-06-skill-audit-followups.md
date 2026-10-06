---
parent: specs/2026-10-06-skill-audit-followups
depth: minimal
closes: [DBT-0084.01, DBT-0084.02]
id: PLN-0100
created: "2026-10-06 07:10:03"
hash: bu9zdov
started: "2026-10-06 07:12:34"
finished: "2026-10-06 07:21:02"
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

- [x] Add `MustNot` `"Ask them now"` and `"Already in isolated workspace"` to `TestSkillBuild`; it fails on today's text.
- [x] In `implementer-prompt.md`, "**Ask them now.** Raise any concerns before starting work." becomes "Stop and report NEEDS_CONTEXT with the question before you start." and the "**While you work:**" lines become "If something is unexpected or unclear, stop and report NEEDS_CONTEXT with the question. Don't guess."; in `SKILL.md`, the two quoted report lines become one line: "Tell the user, in their chat language, the worktree path and branch, or that HEAD is detached and a branch is needed at finish time."
- [x] Commit: `fix(skills): build asks for NEEDS_CONTEXT and reports in the chat language`

### Task 2: debug probes stay out of the checkout

**Files:** `plugin/skills/debug/root-cause-tracing.md`, `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`

**verify:** No debug text tells the agent to add instrumentation or run a probe in the real checkout, and a CI-only probe always ends with the user running it, never a push; list every instrumentation and probe step in the debug folder and where it runs.

- [x] Add `MustNot` `"When you can't trace manually, add instrumentation:"` and `Must` `"never push"` to `TestSkillDebug` (confirm the rule reads the folder's reference files); it fails.
- [x] In `root-cause-tracing.md` the line becomes "When you can't trace manually, instrument a throwaway clone outside the repo:"; in `SKILL.md`, after the CI example's "This reveals" line, add "A probe that can only run in CI: write its steps and ask the user to run them; never push."
- [x] Commit: `fix(skills): debug probes never touch the checkout or push`

### Task 3: version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same `x.y.z` version, one patch above main's at branch time.

- [x] Run `scripts/test ./internal/plugincheck/`; the version check already guards agreement, so the bump has no red test of its own.
- [x] Add 1 to the patch version in all three files.
- [x] Commit: `chore(plugin): bump patch version`

## Polish

### Task 4: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] `plugin/skills/debug/SKILL.md:112`: make it full sentences that keep the example and say to give the steps, for example "**This reveals:** Which layer fails. A probe that only runs in CI: give the user its steps; never push." Stay under the byte cap by tightening this line only; never raise a cap.
- [x] `plugin/skills/debug/root-cause-tracing.md:68`: a full sentence with its condition and verb, for example "If you can't trace it by hand, log in a throwaway clone outside the repo:". Stay under the byte cap; shorten other words on that line only if needed.
- [x] `internal/plugincheck/skill_build_test.go`: add `MustNot` `"ask questions"` so the old implementer line cannot come back.
- [x] Commit: `polish: review notes for PLN-0100`

Polish reverted (e92d59e): the polish review found that `root-cause-tracing.md:68` lost "outside the repo". Every Task 4 item moved to `[debt]`.

## Review notes

- implementer-prompt.md sends anything unexpected to NEEDS_CONTEXT, while BLOCKED may fit a blocking surprise better; the controller treats both the same.
- The ticked tasks quote wording that did not land word for word; 96ff265 shortened it to fit byte caps.
- build/SKILL.md:155 "If the implementer asks questions" is fine; a future "ask questions" pin would collide only if reworded.
