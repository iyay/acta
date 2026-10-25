---
parent: bugs/2026-10-22-build-close-reruns-fast-suite
depth: minimal
closes: [DBT-0139.01]
id: PLN-0195
created: "2026-10-26 06:23:33"
hash: h31qlc6
started: "2026-10-26 06:24:13"
finished: "2026-10-26 06:40:48"
---
# Reuse task test evidence during build close

**Goal:** Resolve BUG-0054 by making build close report existing task test evidence without rerunning a repository-wide suite.
**Spec:** `.acta/specs/2026-10-26-build-close-reuse-test-evidence-design.md`
**Tests:** Focused `scripts/test ./internal/plugincheck -run 'TestSkillBuild|TestBuildClose|TestBudgetFiles'`; `scripts/eval plugin --case build-close-no-suite`; full `scripts/test --full` and existing landing gates.

## Global Constraints

- Follow acta:lean: understand the task and code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build inline in a separate worktree from main only after this plan is approved; keep the Cold read branch intact and never push.
- Change the Close wording and existing build contract checks; the user also approved moving the two existing eval markers into `.git/` and updating grader paths and the fixture path test, with no assertion or threshold removed; keep CLI, runner, commands, settings, and existing budgets unchanged.
- Preserve outstanding type checks, focused checks for later behavior changes, wiki review, progress checks, review, dispatch close, and landing's full suite.
- Treat baseline failures as evidence, not permission to weaken a grader; record counts and native traces under the same client/model settings and require all five candidate trials to pass.
- The intermittent build-tick-failure eval is outside this plan; its baseline and candidate focused rechecks passed.

## Waves

- Wave 1: Task 1.

### Task 1: Separate reporting prior test results from executing checks

**Files:** Modify `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/evals_test.go`, `plugin/evals/build-close-no-suite/scaffold.sh`, and that case's `graders/no-fast-suite.md` and `graders/typecheck-ran.md`.
**Seam:** Existing build-close-no-suite native behavior eval and the existing focused build skill/budget checks.
**verify:** In every candidate trial, close leaves .git/broad-suite-ran absent, creates .git/typecheck-ran, verifies completed task progress, and reaches review without starting it; fixture scripts preserve the same evidence from root and subdirectory, including the existing adversarial path cases; task test results are reported from existing evidence, later behavior changes still require their focused check, and land owns the full suite.
- [x] Red: Run `scripts/eval plugin --case build-close-no-suite --runs 3 --trust-plugin --keep-temp` on the unchanged parent with its binary on PATH, retain the failing marker/native trace evidence, and update only the existing Close contract assertions that currently require the confusing phrase so the focused check fails before the wording change.
- [x] Change: Replace the ambiguous show-tests wording with explicit reuse of recorded task results and outstanding-check execution, preserve the no-second-suite and all other Close gates within current budgets, then run the focused test command and `scripts/eval plugin --case build-close-no-suite --runs 5 --trust-plugin --keep-temp` with the candidate binary on PATH and inspect every native trace for the required behavior.
- [x] Commit: `fix(build): reuse recorded test evidence during close`

## Polish

### Task 2: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] Remove each evidence marker before each working-directory trial so root execution cannot mask a missing subdirectory write.
- [x] Commit: `polish: review notes for PLN-0195`

## Review notes

- Deep full-tier Spec and Standards reviews CLEAN; no BLOCKERs. Polish removes stale markers before every fixture cwd trial; both polish reviews CLEAN. Same revised fixture: unchanged skill failed 3/3, candidate passed 5/5. Focused checks, vet and format passed; isolated no-op-subdirectory mutant passes old test and fails revised test with HOME/TMPDIR pinned. Assertion types and thresholds are unchanged.

## State

### Findings
