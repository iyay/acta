---
parent: specs/2026-10-19-understanding-no-question-design
depth: minimal
id: PLN-0168
created: "2026-10-19 09:40:07"
hash: w2fve66
started: "2026-10-19 09:46:01"
finished: "2026-10-19 09:48:52"
---
# Landing explanation without a question plan

**Goal:** With `review_understanding` on, the land report explains the landed code and asks no question.
**Spec:** `.acta/specs/2026-10-19-understanding-no-question-design.md`
**Tests:** `scripts/test` (fast), `scripts/test --full` (full, land only); task loop: `scripts/test ./internal/plugincheck/ ./internal/setup/ ./internal/cli/ ./internal/config/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The setting keeps its name `review_understanding`; with it false or missing nothing changes.
- Inline in a worktree, English in the repo, never push.

## Waves

- Wave 1: Task 1, Task 2.

### Task 1: Land and review skills drop the question

**Files:** Modify `plugin/skills/land/SKILL.md`, `plugin/skills/review/depth.md`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/evals_test.go`, `plugin/evals/review-understanding-report/` (prompt and grader); delete `plugin/evals/review-understanding-followup/`.
**Seam:** `plugincheck.CheckSkill` with the land and review `SkillRule`s, and the eval case table in `evals_test.go`.
**verify:** No skill or eval text asks the user a question after a land, checks an answer, or offers a scenario; list every place in land, review and the evals that names `review_understanding`, and what each says now. The explanation parts (behavior, example, risk, what tests prove) stay required. The question cannot come back reworded.
- [x] Red: Require the explanation-only wording and ban "optional question", "predicts the outcome" and "optional scenario"; drop the followup eval row; observe `scripts/test ./internal/plugincheck/` fail.
- [x] Change: Rewrite the land Understanding part and depth.md, update the report eval, delete the followup eval; pass the same command.
- [x] Commit: `fix(land): understanding explains the change and asks no question`.

### Task 2: Setting text drops the question

**Files:** Modify `internal/setup/form.go`, `internal/cli/config_cmd.go`, `internal/cli/help.go`, `internal/config/user.go` (comment), and their tests only where they pin the old text.
**Seam:** the setup form option label, the `config set --review-understanding` flag help and `acta help` text.
**verify:** No user-facing text for the setting promises a question; list every surface checked (setup wizard, flag help, help text, config comment) and what each says now.
- [x] Red: Pin the new wording in an existing test that covers that text, or add one in `internal/cli` for the help line; observe `scripts/test ./internal/cli/ ./internal/setup/` fail.
- [x] Change: Reword the four places to "explain the change"; pass `scripts/test ./internal/cli/ ./internal/setup/ ./internal/config/`.
- [x] Commit: `fix(config): review understanding text drops the question`.

## Polish

### Task 3: Review polish

**verify:** Each NOTE below applied, and nothing else changes.
- [x] Report grader header says "before the reply", not "before the question".
- [x] depth.md says the off report has none, not "neither"; the review test pins it.
- [x] Commit: `polish: review notes PLN-0168`

## Review notes

- Round 1: deep, full tier, Spec and Standards reviewers over 6a1004d..7aff979: CLEAN. Polish e8bf354 (grader comment, depth.md "neither") reviewed full tier by both: CLEAN. Notes left: "no offer" in land SKILL.md off-path line is a harmless leftover; the spec's grep check hits the ban-list test strings by design.
