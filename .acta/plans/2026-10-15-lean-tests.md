---
parent: specs/2026-10-15-lean-tests-design
depth: minimal
id: PLN-0144
created: "2026-10-15 10:02:28"
hash: wsfypzr
started: "2026-10-15 10:02:46"
finished: "2026-10-15 10:03:42"
---
# Lean tests Implementation Plan

**Goal:** Slice names a seam per task and counts cases by behavior; tdd tests only at that seam, folds same-path cases into table tests, and goes one test at a time.

**Spec:** `.acta/specs/2026-10-15-lean-tests-design.md`

**Tests:** `scripts/test ./internal/plugincheck`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Skill wording is general and fits every plugin user. English only.
- The tdd folder stays under its 560-line cap.

## Waves

- Wave 1: Task 1, Task 2

### Task 1: Slice names seams and counts cases by behavior

**Files:**
- Modify: `plugin/skills/slice/SKILL.md`
- Test: `internal/plugincheck/skill_slice_test.go`

**verify:** A plan written under the slice skill cannot leave a task without a named seam, and cannot list two cases that run the same code path as separate cases. List every place in the skill (task template, verify rules, minimal depth) that states or shows this.

- [x] Add the new rule phrases to `Must` in `skill_slice_test.go`, run `scripts/test ./internal/plugincheck -run TestSkillSlice`, see it fail, edit the skill, see it pass, commit.

### Task 2: tdd tests at the seam, table tests, one test at a time

**Files:**
- Modify: `plugin/skills/tdd/SKILL.md`, `plugin/skills/tdd/writing-good-tests.md`
- Test: `internal/plugincheck/skill_tdd_test.go`

**verify:** No rule in the tdd folder lets a test reach past the plan's seam, split same-path input cases into separate tests, or write every test before any code; test names use glossary words when `.acta/wiki/glossary.md` exists; refactor stays in the loop. List each rule and where it lives.

- [x] Add the new rule phrases to `Must` in `skill_tdd_test.go`, run `scripts/test ./internal/plugincheck -run TestSkillTDD`, see it fail, edit the two files, see it pass, commit.

## Polish

### Task 3: Review polish

**verify:** Each NOTE below applied, and nothing else changes.

- [x] Minimal plans: the spec's yes covers their seams (slice).
- [x] tdd falls back to the public interface callers use when no seam is named.
- [x] "Never write all the tests before any code" replaces the ambiguous wording.
- [x] Guards for the Seam field in the template and the minimal shape; minimal line says about 7 lines a task.
- [x] Commit: `polish: review notes PLN-0144`

## Review notes

- Round 1: two reviewers (deep), CLEAN; polish review CLEAN.
- budget_test.go byte caps raised to the new file sizes; not in the plan's Files, needed for the new rule text.
- Minimal shape's one box for all failing tests sits a little at odds with "one test, then its code"; not a contradiction.
- plugin/evals/lean-tests graders not re-run after the table-test rule.
