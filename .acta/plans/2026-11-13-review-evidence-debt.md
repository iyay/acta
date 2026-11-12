---
parent: debt/2026-11-13-review-evidence
depth: minimal
closes: [DBT-0165.01, DBT-0165.02, DBT-0165.03]
after: [PLN-0237]
id: PLN-0238
created: "2026-11-13 06:13:07"
hash: f3qekfa
---
# Review keeps incomplete ranges blocked

**Goal:** Incomplete reviews retain their ranges and verdict tests measure Coverage recovery and an executed repro.

**Spec:** `.acta/specs/2026-11-13-workflow-debt.md` (SPC-0208)
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build inline, serially, in a worktree under `~/.wwt/acta/`; never push.
- Keep source text generic across repos and harnesses; adjust byte caps only where needed, without cutting safety instructions.
- Validate changed graders with passing and failing samples before scoring a skill; use `evalomp.LoadCases` and `evalomp.Grade` in plugincheck for fixed graders, and measure judge rubrics on planted traces with the existing eval runner.
- For every skill-text task, run each named case once per call with `scripts/eval --case <name> --runs 10 -j 8` on old and new text under identical corrected graders and scaffolds; use an isolated copy for the old arm and the branch binary on PATH, never replace the installed binary.
- Record raw results with `acta state set`; isolate and exclude any regressing text change and keep its debt open, removing that item from `closes:` before land.
- Full suites run only at land; focused checks run after each change. Run scaffolds and mutation checks with fresh HOME and TMPDIR outside the repo.

## Build brief

- Land order: start only after PLN-0237 lands on main; these serial plans share plugincheck and eval files.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Grade Coverage recovery and executed false-BLOCKER checks

**Files:** `plugin/evals/review-clean-has-evidence/{prompt.md,scaffold.sh,graders/brief-asks-coverage.md,graders/verdict.md}`, new `plugin/evals/review-clean-has-evidence/graders/coverage-recovery.md`, `plugin/evals/review-false-blocker/{prompt.md,scaffold.sh,graders/clean-verdict.md}`, new `plugin/evals/review-false-blocker/graders/repro-ran.md`, `internal/plugincheck/workflow_debt_evals_test.go`, `internal/plugincheck/evals_test.go`
**Seam:** loaded eval graders against planted reports, tool traces and scaffold repro markers.
**verify:** A brief that mentions Coverage alone cannot pass the recovery check: a report lacking it must be returned to the same reviewer once, and a second omission leaves the axis BLOCKED. The false-BLOCKER case passes only after the real repro executes successfully; reading or quoting its command is insufficient. Accept a rejected finding with case variants and emphasized CLEAN in the orchestrator's own final verdict, while rejecting quoted-only verdicts and an unexecuted repro. Validate each fixed grader with positive and negative samples and each judge against planted traces.
- [ ] Red: add `TestWorkflowDebtReviewGraders` and missing/repeated Coverage scenarios, then run `scripts/test ./internal/plugincheck -run 'TestWorkflowDebtReviewGraders|TestEvalCases|TestEvalLLM'` to show that the current graders accept stand-ins and reject valid verdict forms.
- [ ] Change: make the scaffold expose a runnable repro marker and incomplete reviewer reports, replace proxy checks with observed recovery/repro grading, allow only the needed judge rubric in the existing allowlist, and run both cases on unchanged skill text to establish corrected baselines.
- [ ] Commit: test(review): grade Coverage recovery and executed repros

### Task 2: Preserve incomplete ranges through all three rounds

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`, new `plugin/evals/review-incomplete-rounds/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`, `.acta/wiki/review-base-merge-base.md`
**Seam:** review/build required rules and `review-incomplete-rounds`, with Task 1 cases as evidence controls.
**verify:** A cut list or missing Coverage never becomes CLEAN solely because listed BLOCKERs were rejected. A mixed round retains both its fixes and every incomplete range in the next review; an incomplete-only round repeats its range without inventing a fix task. Round 3 still incomplete stops for the user and never starts round 4. The build handoff and review select the same merge-base range even when main moved. List each round's input range and decision in planted mixed, incomplete-only and third-round traces.
- [ ] Red: add focused `TestReviewIncompleteRanges` and `TestBuildReviewMergeBase` checks plus the incomplete-rounds case, then run `scripts/test ./internal/plugincheck -run 'TestReviewIncomplete|TestBuildReviewMergeBase|TestReviewVerdictNeedsEvidence|TestEvalCases'` and the case on old text to expose the lost ranges and contradictory CLEAN path.
- [ ] Change: state the cumulative incomplete-range rule and round-3 stop, fix build's range handoff, update the wiki, then run the focused checks and measure `review-incomplete-rounds`, `review-clean-has-evidence`, `review-false-blocker` and `build-close-no-suite` for 10 runs per arm.
- [ ] Commit: fix(review): preserve incomplete ranges within the round budget

