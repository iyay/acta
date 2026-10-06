---
parent: specs/2026-10-06-skill-prompt-audit-fixes
depth: minimal
id: PLN-0099
created: "2026-10-06 06:29:46"
hash: nv3bej2
started: "2026-10-06 06:32:45"
finished: "2026-10-06 06:41:18"
---
# skill prompt audit fixes Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Skill text names no missing command, no one-user setup, and no rule that another rule in the plugin contradicts.

**Spec:** `.acta/specs/2026-10-06-skill-prompt-audit-fixes.md`

**Tests:** `scripts/test ./internal/plugincheck/`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Byte caps in `internal/plugincheck/budget_test.go` stay as they are; a rewrite that grows a file past its cap is shortened, never the cap raised.
- Each new `MustNot` pin is the stale phrase verbatim, so it fails before the text change and passes after.
- The exact new text for each hunk is in `/private/tmp/claude-501/-Users-iyay-Nayakatara-acta/a872a391-3805-4cd2-bd9c-b54d64efc29d/scratchpad/prompt-audit.diff`; quoted text below wins where they differ.

## Waves

- Wave 1: Tasks 1-7 (each owns its own skill files and its own test file).
- Wave 2: Task 8.

### Task 1: migrate names no missing command

**Files:** `plugin/skills/migrate/SKILL.md`, `internal/plugincheck/skill_migrate_test.go`

**verify:** No text in the migrate skill tells an agent to run an acta subcommand that `internal/cli/cli.go` does not dispatch; list every acta command the skill names.

- [x] In `TestSkillMigrate`, drop `"acta migrate superpowers"` and `"--apply"` from `Must`, add `"git mv"` to `Must` and `"acta migrate superpowers"` to `MustNot`; it fails because the skill still names the command.
- [x] Replace the `## superpowers docs` body with one paragraph: superpowers docs take the route below, each file moved with `git mv` instead of written new, and after the commit `docs/superpowers` is dropped from `legacy` in `.acta.yaml`.
- [x] Commit: `fix(skills): migrate names no missing acta command`

### Task 2: review text agrees with itself and its reviewer template

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/review/code-reviewer.md`, `internal/plugincheck/skill_review_test.go`

**verify:** Every rule the reviewer template gives agrees with the review skill's three questions, deep lens limits and output format, and the skill states one review rule per kind of commit; list each rule checked.

- [x] In `TestSkillReview`, change the pin to `"The polish commit counts as no round. Every polish commit gets the two reviewers."` and add `MustNot` `"explicit instruction-file violation"`, `"Acknowledge strengths"`, `"Production readiness"`; check that the rule reads `code-reviewer.md` (add a read of it if not) so it fails first.
- [x] Apply: `Every commit gets` becomes `Every polish commit gets`; `(explicit instruction-file violation)` becomes `(performative, says nothing)`; the `## What to Check` checklist becomes "Answer the review skill's three questions for your axis. Use the deep lens (trace every user value to its sink, check every caller, adversarial inputs, silent error swallowing, cross-tenant access) only where the diff touches a trust boundary, auth, money, a migration or a delete."; drop the `Acknowledge strengths` line.
- [x] Commit: `fix(skills): review template follows the review skill`

### Task 3: build and its implementer prompt agree

**Files:** `plugin/skills/build/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, `internal/plugincheck/skill_build_test.go`

**verify:** Build text never lets work run in the main checkout, never states one repo's setup as a fact, and the implementer template agrees with the build skill on models, the brief and TDD; list every place in both files that names the main checkout, a model or TDD.

- [x] In `TestSkillBuild`, add `MustNot` `"is a symlink to an untracked"`, `"explicit consent"`, `"working in the current directory instead"`, `"### 5. Complete"`, `"[BRIEF_FILE]"`, `"TDD if required"`, `"and only when\n         `"`; they fail on today's text (confirm the rule reads `implementer-prompt.md`, add the read if not).
- [x] Apply in `SKILL.md`: house rules line becomes "Copy in the house rules too: any `CLAUDE.md` or `AGENTS.md` git does not track (file or symlink). A fresh worktree holds only tracked files, so without them the agent runs the ticket with no house rules."; sandbox fallback becomes "if `git worktree add` fails with a permission error, stop and tell the user the sandbox blocked it. Never work in the main checkout instead."; `## Setup` keeps only "Work happens only in the worktree from ## Worktree."; `### 5. Complete the task` becomes `### 3. Complete the task`. In `implementer-prompt.md`: the model line says `sonnet` in Claude Code only under `subagent_models: split`, else no model, and on omp always `agent="task"`; `[BRIEF_FILE]` lines become "Your task is Task N in [PLAN_PATH]. Read only that task's section, never the whole plan. The exact values to use are below."; `Did I follow TDD if required?` becomes `Did every change start from a failing test I watched fail?`.
- [x] Commit: `fix(skills): build and implementer prompt agree`

### Task 4: slice header and overview match plan depth

**Files:** `plugin/skills/slice/SKILL.md`, `internal/plugincheck/skill_slice_test.go`

**verify:** No slice rule demands of a minimal plan what Plan depth says it lacks, and no slice rule asks for chat narration or more than one commit per task; list each such rule.

- [x] Add `MustNot` `"Every plan MUST start"`, `"Announce at start"`, `"questionable taste"`, `"Frequent commits"`; it fails.
- [x] Apply: header line becomes "**Every `full` plan starts with this header** (a `minimal` plan uses the shorter one in Plan depth):"; drop the Announce line; Overview becomes "Write the plan for a skilled implementer who knows nothing about this codebase, its tools or its domain, and who sees only their own task. Each task names the files to touch, the code, the tests and how to run them. Bite-sized tasks. DRY. YAGNI. TDD. One commit per task."
- [x] Commit: `fix(skills): slice header and overview match plan depth`

### Task 5: debug phases 1 to 3 change no checkout file

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`

**verify:** No step in phases 1 to 3 tells the agent to change a file in the checkout; list every step that writes anything and where it writes.

- [x] Add `MustNot` `"add diagnostic instrumentation"`, `"SMALLEST possible change"`; it fails.
- [x] Apply: `**BEFORE proposing fixes, add diagnostic instrumentation:**` becomes `**BEFORE proposing fixes, instrument a scratch copy:**`; `Make the SMALLEST possible change to test hypothesis` becomes `Test it with the SMALLEST probe, in a scratch copy`.
- [x] Commit: `fix(skills): debug probes stay out of the checkout`

### Task 6: tdd has one rule on exemptions and test scope

**Files:** `plugin/skills/tdd/SKILL.md`, `internal/plugincheck/skill_tdd_test.go`

**verify:** The tdd skill gives one answer on config exemptions and one on testing code the plan does not cover; list every line that rules on either.

- [x] Add `MustNot` `"- Configuration files"`, `"## Final Rule"`, `"Add tests for existing code"`; it fails.
- [x] Apply: drop the `- Configuration files` exception; the "Existing code has no tests" row reality becomes "Test the behavior you change. Tests for the rest are a note for a later plan."; drop the `## Final Rule` section.
- [x] Commit: `fix(skills): tdd rules on exemptions and scope agree`

### Task 7: shape review gate speaks the user's language

**Files:** `plugin/skills/shape/SKILL.md`, `internal/plugincheck/skill_shape_test.go`

**verify:** No shape step gives a fixed English sentence to say to the user; list every quoted chat line in the skill.

- [x] Add `MustNot` `"Please review it and let me know"`; it fails.
- [x] Apply: the gate becomes "**User review gate.** In the user's chat language, give the spec path, say it is committed, and ask for a review before the plan. Then wait."
- [x] Commit: `fix(skills): shape review gate uses the chat language`

### Task 8: version 0.1.21

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same `x.y.z` version, one patch above main's.

- [x] Run `scripts/test ./internal/plugincheck/`; it passes on 0.1.20 so the bump has no red test of its own (version check already guards agreement).
- [x] Change 0.1.20 to 0.1.21 in all three files.
- [x] Commit: `chore(plugin): version 0.1.21`

## Polish

### Task 9: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] `plugin/skills/debug/SKILL.md` Phase 1 step 4 and Phase 3 probe: "a scratch copy" becomes "a throwaway clone outside the repo" ("scratch" already means acta's Scratchpad, and a copy inside the checkout is a file change). Stay under the debug byte cap; shorten other words on those two lines if needed, never the cap.
- [x] `plugin/skills/build/implementer-prompt.md`: "The exact values to use are below." becomes "The exact values to use are in Context below." so it names the section that holds them.
- [x] Commit: `polish: review notes for PLN-0099`

## Review notes

- review/SKILL.md:34 deep lens says "error swallowing" (trimmed for the byte cap) while code-reviewer.md says "silent error swallowing".
- Byte caps were not lowered for files that shrank (code-reviewer.md, implementer-prompt.md, migrate); the spec kept caps as they are.
- skill_build_test.go:59 pin "and only when\n         " depends on exact wrapping; the new model line has no Must pin.
- implementer-prompt.md "The exact values to use" now points at Context, which holds scene-setting text, not a values slot.
- Commit 7fed3c9 alone fails TestBudgetFiles; e927ef6 fixes it, the range is green.
- debug/SKILL.md polish: "BEFORE proposing fixes" became "First", "SMALLEST probe" became "one probe"; meaning kept by Phase 1 and the lines below.
