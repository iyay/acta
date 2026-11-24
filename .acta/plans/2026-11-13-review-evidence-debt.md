---
parent: debt/2026-11-13-review-evidence
depth: minimal
closes: [DBT-0165.01, DBT-0165.02, DBT-0165.03]
after: [PLN-0237]
id: PLN-0238
created: "2026-11-13 06:13:07"
hash: f3qekfa
started: "2026-11-13 18:44:11"
finished: "2026-11-13 19:24:33"
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
- [x] Red: add `TestWorkflowDebtReviewGraders` and missing/repeated Coverage scenarios, then run `scripts/test ./internal/plugincheck -run 'TestWorkflowDebtReviewGraders|TestEvalCases|TestEvalLLM'` to show that the current graders accept stand-ins and reject valid verdict forms.
- [x] Change: make the scaffold expose a runnable repro marker and incomplete reviewer reports, replace proxy checks with observed recovery/repro grading, allow only the needed judge rubric in the existing allowlist, and run both cases on unchanged skill text to establish corrected baselines.
- [x] Commit: test(review): grade Coverage recovery and executed repros

### Task 2: Preserve incomplete ranges through all three rounds

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`, new `plugin/evals/review-incomplete-rounds/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`, `.acta/wiki/review-base-merge-base.md`
**Seam:** review/build required rules and `review-incomplete-rounds`, with Task 1 cases as evidence controls.
**verify:** A cut list or missing Coverage never becomes CLEAN solely because listed BLOCKERs were rejected. A mixed round retains both its fixes and every incomplete range in the next review; an incomplete-only round repeats its range without inventing a fix task. Round 3 still incomplete stops for the user and never starts round 4. The build handoff and review select the same merge-base range even when main moved. List each round's input range and decision in planted mixed, incomplete-only and third-round traces.
- [x] Red: add focused `TestReviewIncompleteRanges` and `TestBuildReviewMergeBase` checks plus the incomplete-rounds case, then run `scripts/test ./internal/plugincheck -run 'TestReviewIncomplete|TestBuildReviewMergeBase|TestReviewVerdictNeedsEvidence|TestEvalCases'` and the case on old text to expose the lost ranges and contradictory CLEAN path.
- [x] Change: state the cumulative incomplete-range rule and round-3 stop, fix build's range handoff, update the wiki, then run the focused checks and measure `review-incomplete-rounds`, `review-clean-has-evidence`, `review-false-blocker` and `build-close-no-suite` for 10 runs per arm.
- [x] Commit: fix(review): preserve incomplete ranges within the round budget

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] Include pending middle fix range7777777..8888888 in the third-history round3 input and rubric, alongside6666666..7777777 and8888888..9999999. Calibrate rejection of omission and remeasure the affected case old/new under identical corrected fixtures; production policy stays unchanged.
- [x] Commit: `polish: review notes for PLN-0238`

## Review notes

- Round 1, auto selected standard, full tier because agent instructions changed: input e9291623b1c9a8e01cad43e68955e19f134033b6..49555669b05ef163d6410ef3da35d875d8ce6fc2. Both axes CLEAN, complete Coverage. BLOCKER 0, NOTE 1, pre-existing 0 after merging duplicate findings. One [fix] measurement NOTE: third-history fixture/rubric omits pending middle fix7777777..8888888; fix via one polish task, production policy unchanged. Package PASS Spec14.651s/Standards13.978s; isolated production policy/grader reversion RED. No unexpected worktrees or undispatched code commits.
- Review: auto selected standard (full tier: agent instructions); Round 1 CLEAN; BLOCKER 0, NOTE 1 (fixed), pre-existing 0; polish CLEAN.
- Polish input49555669b05ef163d6410ef3da35d875d8ce6fc2..f48fc059fb09b205bf2cfda9174363e99ca2bd03, no round increment. Spec/Standards complete Coverage, no new findings; narrow checks1.422s/1.427sPASS. Both verified actual native RED/GREEN rubric identity; corrected paired targetold10/10new10/10, unchanged othercontrols. Accepted NOTE resolved. No unexpected worktrees or undispatched code commits. No secondpolish or newdebt.

## State

### Next

All tasks/reviews/gates complete: done3/3, main and polish CLEAN, full28packages/0flakes and selected23/23 first-run eval PASS. Proceed tidy main review-evidence-debt with no fold(parent has feature commits). Check tidy-tip diff and reusegates only same/planning tree; main must be clean/unchanged before separate move.
After main moves: specdates/hash check, final plan done3/3 and a Task1/2/3commit ancestor of HEAD; then plain worktree removal, permitted branch-D and tidyref cleanup. Close DBT0165.01-.03 with tickall and commitdebtfile. Do not push/attach.
Then approvedPLN0239 lsr8dl1 starts inline in new worktree under~/.wwt/acta after238land. Plan already read: oneTask corrects two reviewer-role evals; no skilltextchanges, so no slowgate triggered by its plan paths, but required namedcases10 each and fullsuite remain. Continue approved0240-0241serially, no reapproval.

### Findings

All three tasks committed and done, including one polish. Main review and polish both CLEAN with full Coverage; totals BLOCKER0 NOTE1(fixed) pre-existing0. Exact review ranges/level are in Review notes. No unexpected worktrees or source changes after review.
Watched RED/green for old proxy/verdict graders and merge-base/cumulative-range contracts. Narrow packages14.055s/13.799s PASS; final budget/property checks0.159s PASS. Standards isolated production reversion RED. Real Git with advanced main proves the actual build assignment returns branch point, not parent tip.
Native Coverage judge9/9 expected; rounds judge7/7 beforepolish. Polish beforegrader accepted missing middle fix (watched RED); corrected rubric8/8 expected2validPASS6invalidFAIL. Probes/raw in review-debt-probes-20261113,review-rounds-probes-20261113,review-polish-probes-20261113. Setup failures and prior false negatives retained and excluded from skill-effect claims.
Final corrected identical-fixture pairs40/40each: allfour namedcases10old/10new, policy hashes and aggregates in review-evidence-final-bin-20261113/paired-summary.json. Corrected three-range targetold10/10 tmp.WGcxZxZZCu,new10/10 tmp.2dDHBrnRji. Earlier sourcecandidates1/2target9/10 retained; finalthird candidate has no measured regression, no fourthsourcefix or guarantee claim.
Fresh full suite28packages,0flakes,exit0 testedc0125b7e437e840faa50119f7f3015522b370619 treee7e785fcec6367fbee5d93c32f75a17ac73a4c27. Complete log .git/acta-baseline-20261113T123733Z-32617.log; PASS baseline atomically published. Eval --tag=review --tag=build23/23 first-run PASS,exit0,0retries/flakes; log .git/acta-review-evidence-eval-20261113T123733Z-32616.log, reporttmp.DfveXTj2j3. No nonplanning changes since testedsource.
Wiki review-base-merge-base updated once, correct enduring range rule. Wiki gate only stale warnings/no errors. Remaining findings above are archived evidence/limitations, not new unresolved work. No push/attachments; all original approvals remainvalid.
