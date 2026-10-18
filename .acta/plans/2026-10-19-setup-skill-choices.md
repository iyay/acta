---
parent: specs/2026-10-19-setup-skill-choices-design
depth: minimal
id: PLN-0162
created: "2026-10-19 05:36:20"
hash: golks4g
started: "2026-10-19 05:38:11"
finished: "2026-10-19 05:42:07"
---
# Setup skill choices plan

**Goal:** In Claude Code, `acta:setup` asks the settings itself as `AskUserQuestion` choices instead of sending the user to the wizard.
**Spec:** `.acta/specs/2026-10-19-setup-skill-choices-design.md`
**Tests:** `scripts/test` (fast), `scripts/test --full` (full, land only); task loop: `scripts/test ./internal/plugincheck/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- No Go product code changes. Skill text names settings, never `acta config set` flags (the existing `MustNot` list stays).
- Options stay neutral for every user; no wording shaped on one user's setup. Plugin installs stay in the wizard only.
- Build inline in an isolated worktree, repo content in English, never push.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2.

### Task 1: Setup skill asks with AskUserQuestion

**Files:** Modify `plugin/skills/setup/SKILL.md`, `internal/plugincheck/skill_setup_test.go`, `internal/plugincheck/budget_test.go`; create `plugin/skills/setup/choices.md` only if the text does not fit in `SKILL.md`.
**Seam:** `plugincheck.CheckSkill` with the setup `SkillRule`, and the budget tables in `budget_test.go`.
**verify:** Every setup path in the skill text is covered and none contradicts another: Claude Code with `AskUserQuestion` (first run asks every core setting; run again asks only what plain `acta config show` marks missing or `(default)`), one-setting change with `overrides` from `--json`, and no `AskUserQuestion` (wizard, then chat fallback). List each path and the line that owns it. No call asks more than 4 questions or 4 options; the acta block rules are unchanged.
- [x] Red: Add `AskUserQuestion` and `acta config show` (plain, with `(default)`) to `Must`, raise `MaxLines` and budgets to the planned size; observe `scripts/test ./internal/plugincheck/ -run 'TestSkillSetup|Budget'` fail.
- [x] Change: Rewrite the skill text per the spec; pass the same command.
- [x] Commit: `feat(setup): ask setup settings as choices in Claude Code`.

### Task 2: Eval case for the setup fallback

**Files:** Create `plugin/evals/setup-fallback-no-choices/` (`case.yaml`, `prompt.md` with `allowed_tools` including `AskUserQuestion`, `graders/`), following the layout of `plugin/evals/second-brainstorm-choices/`.
**Seam:** the eval case as `claude plugin eval` runs it, plus the plugincheck eval layout tests.
**verify:** With no `AskUserQuestion` (the eval sandbox never grants it), the case fails when setup skips `acta config show`, writes a setting nobody picked, or leaves no way to finish setup, and passes on the Task 1 skill. List the grader checks and what each rejects. Layout checks pass with `scripts/test ./internal/plugincheck/`; the case itself runs alone through `scripts/eval` filtered to this case, with the branch binary on PATH.
- [x] Red: Write the case and graders; run it against main's skill text and observe it fail.
- [x] Change: Run it on the branch; observe pass, and pass `scripts/test ./internal/plugincheck/`.
- [x] Commit: `test(evals): setup falls back when choices are missing`.

## State

### Next

Land.

### Findings

Recorded in plugin/evals/FACTS.md (2026-10-19 entry).

## Fix round 1

### Task 3: Fix round 1

**verify:** The one-setting override flow, the executor answer, the wiki offer and the acta block rules apply on every path, not only without AskUserQuestion; the chat path covers tone, subagent models and plugin installs; the fallback path tells the agent to read the config.
- [x] BLOCKER: move the override, executor, wiki and block paragraphs under `## Every path` (both reviewers, SKILL.md:20).
- [x] [fix] doctor reports a missing plugin: point to `! acta setup` for the install only.
- [x] [fix] first run also asks tone and subagent models; later runs ask what is `(default)` or not printed.
- [x] [fix] acta block rule is one rule for every path outside the wizard.
- [x] [fix] fallback path runs steps 1 and 2 first; grader renamed `reads-config`.
- [x] Commit: `fix(setup): rules that apply on every path`

## Polish

### Task 4: Review polish

**verify:** Each NOTE below applied, and nothing else changes.
- [x] Fallback path runs doctor and config show read-only; it asks nothing before the wizard.
- [x] Acta block rule limited to full setup runs, so a one-setting change never creates a CLAUDE.md.
- [x] Chat fallback asks the same settings as step 2, not "marked missing".
- [x] Later runs do not re-ask a skipped tone or the theme.
- [x] Commit: `polish: review notes PLN-0162`

## Review notes

- Round 1: deep, full tier, Spec and Standards reviewers over 501261e..7cfe99e: BLOCKED (override flow scoped under No AskUserQuestion); fix round 1 ace3a72. Round 2 over 7cfe99e..ace3a72: CLEAN. Polish 5e52b8e reviewed CLEAN. Two (medium) NOTEs filed as debt; notes left: description still says "optional" block, "full setup run" undefined, redundant --plan-split clause.
