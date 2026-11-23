---
parent: specs/2026-11-23-fewer-tests-design
depth: minimal
id: PLN-0252
created: "2026-11-23 20:40:14"
hash: h3b0azk
---
# Fewer, leaner tests

**Goal:** acta skills make agents add a row before a new test func, block text tests in review, and audit a repo's tests on request.

**Spec:** `.acta/specs/2026-11-23-fewer-tests-design.md` (SPC-0218).
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched and `scripts/eval --case <name>` for one case; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches `plugin/skills/tdd/`, `slice/`, `debug/`, `review/` and `cleanup/`, so the gate runs as `scripts/eval --tag=tdd`, `--tag=slice`, `--tag=debug`, `--tag=review` and `--tag=cleanup`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Replace old skill text where the new rule takes over; add only what is new. A cap in `internal/plugincheck` goes up only by what the edit measures it needs.
- No new `Must` phrase in any plugincheck skill test: skill text is proven by the eval case of its task.
- Each task's red is its own eval case run alone with `scripts/eval --case <name>`, never the whole gate.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3
- Wave 4: Task 4
- Wave 5: Task 5

(One task per wave: every task may touch `internal/plugincheck/budget_test.go`.)

### Task 1: tdd RED step adds a row first

**Files:** `plugin/skills/tdd/SKILL.md`, `plugin/evals/tdd-row-not-func/` (new), `internal/plugincheck/budget_test.go`
**Seam:** the tdd RED step text, proven by `scripts/eval --case tdd-row-not-func`
**verify:** A red test for a unit that has a table test is a row in it, seen failing first, with no new test func; the step names the order row, then assert, then new func, and points to "Behavior, not text". Name the grader checks: row added, red run shown before the change, test func count unchanged.
- [ ] Red: `scripts/eval --case tdd-row-not-func` fails on the current text because the agent adds a new test func.
- [ ] Change: replace the "Write one minimal test" sentence with the three-step order and the pointer; raise the tdd byte cap only by what it needs.
- [ ] Commit: feat(tdd): a red test is a row in an existing table first

### Task 2: slice tasks name the row

**Files:** `plugin/skills/slice/SKILL.md`, `plugin/evals/slice-row-not-func/` (new), `internal/plugincheck/budget_test.go`, `internal/plugincheck/skill_slice_test.go`
**Seam:** the slice task rule, proven by `scripts/eval --case slice-row-not-func`
**verify:** A plan task for a unit with a table test names that test and the row it adds; a new test func appears only with a one-line reason. Name the grader checks: test file and table named, no new func without a reason line.
- [ ] Red: `scripts/eval --case slice-row-not-func` fails on the current text because the task names a new test func.
- [ ] Change: add the rule to the task section of slice SKILL.md, replacing text where it overlaps; raise byte and line caps only by what they need.
- [ ] Commit: feat(slice): plan tasks name the existing test and row

### Task 3: debug regression test follows the order

**Files:** `plugin/skills/debug/SKILL.md`, `plugin/evals/debug-regression-row/` (new), `internal/plugincheck/budget_test.go`, `internal/plugincheck/skill_debug_test.go`
**Seam:** debug phase 4 step 1, proven by `scripts/eval --case debug-regression-row`
**verify:** The regression test proposed for a bug in a unit with a table test is a row in that table. Name the grader check.
- [ ] Red: `scripts/eval --case debug-regression-row` fails on the current text because the proposed test is a new func.
- [ ] Change: phase 4 step 1 says the regression test follows the tdd order; raise caps only by what they need.
- [ ] Commit: feat(debug): regression test is a row in an existing table first

### Task 4: Review blocks text tests

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/review/code-reviewer.md`, `plugin/evals/review-planted-bug/` (scaffold, prompt, a new grader), `internal/plugincheck/budget_test.go`
**Seam:** the review finding bar, proven by `scripts/eval --case review-planted-bug`
**verify:** A test the diff adds that asserts a code, script, skill or config file contains an exact line or phrase is a BLOCKER with its `file:line` and the path it reads; it is the one test finding that is a BLOCKER, and both finding bars and the "Always a NOTE" list agree. The planted `src/discount.py:4` BLOCKER is still found.
- [ ] Red: with the scaffold's new text test and the new grader, `scripts/eval --case review-planted-bug` fails because the reviewer calls the text test a NOTE.
- [ ] Change: add the exception next to the BLOCKER definition in both files and name it in the NOTE lines; drop "The repo has no tests" from the prompt; raise caps only by what they need.
- [ ] Commit: feat(review): a text test is a BLOCKER

### Task 5: Test audit mode in cleanup

**Files:** `plugin/skills/cleanup/SKILL.md`, `plugin/skills/cleanup/tests.md` (new), `plugin/evals/cleanup-test-audit/` (new), `internal/plugincheck/budget_test.go`, `internal/plugincheck/skill_cleanup_test.go`
**Seam:** the cleanup skill on "audit the tests", proven by `scripts/eval --case cleanup-test-audit`
**verify:** On a test audit ask the agent loads cleanup, skips the planning steps, reports the text test and the near duplicate pair with `file:line`, leaves the sound test out, deletes no file, runs no `acta cleanup`, and asks before any `acta debt new`. Name each grader check.
- [ ] Red: `scripts/eval --case cleanup-test-audit` fails on the current text because the agent runs `acta cleanup` or edits tests.
- [ ] Change: new description, the skip line before step 1, `tests.md` per the spec; add the `fileCaps` entry for `tests.md` and raise the description and folder caps only by what they need.
- [ ] Commit: feat(cleanup): audit a repo's tests for text tests and duplicates
