---
parent: specs/2026-10-19-setup-skill-choices-design
depth: minimal
id: PLN-0162
created: "2026-10-19 05:36:20"
hash: golks4g
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
- [ ] Red: Add `AskUserQuestion` and `acta config show` (plain, with `(default)`) to `Must`, raise `MaxLines` and budgets to the planned size; observe `scripts/test ./internal/plugincheck/ -run 'TestSkillSetup|Budget'` fail.
- [ ] Change: Rewrite the skill text per the spec; pass the same command.
- [ ] Commit: `feat(setup): ask setup settings as choices in Claude Code`.

### Task 2: Eval case for setup choices

**Files:** Create `plugin/evals/setup-asks-choices/` (`case.yaml`, `prompt.md` with `allowed_tools` including `AskUserQuestion`, `graders/`), following the layout of `plugin/evals/second-brainstorm-choices/`.
**Seam:** the eval case as `claude plugin eval` runs it, plus the plugincheck eval layout tests.
**verify:** The case fails when the skill tells the user to run `! acta setup` instead of asking, or asks without `AskUserQuestion`, and passes on the Task 1 skill. List the grader checks and what each rejects. Layout checks pass with `scripts/test ./internal/plugincheck/`; the case itself runs alone through `scripts/eval` filtered to this case, with the branch binary on PATH.
- [ ] Red: Write the case and graders; run it against main's skill text and observe it fail.
- [ ] Change: Run it on the branch; observe pass, and pass `scripts/test ./internal/plugincheck/`.
- [ ] Commit: `test(evals): setup asks settings as choices`.
