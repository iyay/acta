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

- [ ] [fix] Include pending middle fix range7777777..8888888 in the third-history round3 input and rubric, alongside6666666..7777777 and8888888..9999999. Calibrate rejection of omission and remeasure the affected case old/new under identical corrected fixtures; production policy stays unchanged.
- [ ] Commit: `polish: review notes for PLN-0238`

## Review notes

- Round 1, auto selected standard, full tier because agent instructions changed: input da73a2e139da583f9b96c7b76c123766694b3aab..49555669b05ef163d6410ef3da35d875d8ce6fc2. Both axes CLEAN, complete Coverage. BLOCKER 0, NOTE 1, pre-existing 0 after merging duplicate findings. One [fix] measurement NOTE: third-history fixture/rubric omits pending middle fix7777777..8888888; fix via one polish task, production policy unchanged. Package PASS Spec14.651s/Standards13.978s; isolated production policy/grader reversion RED. No unexpected worktrees or undispatched code commits.

## State

### Next

Main review round1 overda73a2e..4955566 CLEAN both axes, complete Coverage; merged totals BLOCKER0 NOTE1 pre-existing0. Same [fix] NOTE from both: third-history fixture/rubric omitted middle pending fix7777777..8888888. Production policy correct and unchanged. Review notes recorded exact range; no unexpected worktrees/code commits.
Polish task3 ACTIVE (acta plan add polish, tickstart done). Native pre-polish test RED: old grader passes a sample dropping middle range, expectedFalse, probe before-native.log. Two fixture/rubric files now include all3 pending ranges. Narrow case/schema/scaffold guard tests1.484sPASS, diffcheck clean. Native aftercalibration8samples(2validPASS6invalidFAIL expected), session26889 logreview-polish-probes-20261113/after-native.log; await actualmatch counts.
Final private binary building sessionawaittool, directoryreview-evidence-final-bin-20261113. Aftercalibration, capture oldpolicy+corrected3rangefixture inordinaryprivatecopy and run affectedtargetold/new10each under samegrader; other3 controls/sourceinstructions unchanged so no repeatedruns. Originalcandidate3 beforepolish40/40each retained; supersedeonlytargetcase measurement.
Then commit onepolish with TaskPLN-f3qekfa#3, tick/readback/wavecommit, and two fresh read-only reviewers on polish range because graderinstructions/fulltier. Polishcountsno reviewround. No secondpolish; its notes become debt/note. CLEAN then fullsuite and union --tag=review --tag=build (23 cases) with finalmatchingbinary; retain actualstatuses/logs/onlyPASSbaseline.
Aftergates, tidy land main, final done3/3 and per-task ancestor proof beforecleanup; closeall3DBT0165 items. Continueapproved239-241inline serially, no reapproval/push/docsattachments.

### Findings

Both tasks committed/ticked/read back done2/2: Task1 a0d56f9,Task2 69b2045; wiki c787ee2. Watched RED preceded each change. Focused package Task1 14.055s and Task2 guard/corrected assertions13.799s PASS; final budget/range/Git property checks0.159s PASS. gofmt/diffcheck clean; full suite waits for land.
Task1 grader controls validate quoted-only rejection, case/bold own verdicts, actual delivery/repro markers and missing/extra calls. Native Coverage calibration9/9 expected(3validPASS6invalidFAIL), tmp.ZGUhG08BYv. Unsupported asyncAgent follow-up pilots were setup failures, replaced with deterministic same-reviewer transport and actual delivery records. Repro marker proves execution, not general clone compliance. All raw probes retained in review-debt-probes-20261113.
Precorrection baselines9/10 each were false negatives: unexecuted command after failed compound input wrongly counted as delivered; real false/dropped finding rejected because literalwordreject absent. Corrected graders validated before final measurement. Original raw snapshots/logs retained, never counted as instruction improvements.
Final unchanged-fixture old/new comparisons40/40 each: Coverage10/10,repro10/10,buildclose10/10,incomplete-ranges10/10. Summary/policy hashes review-evidence-candidate3-bin-20261113/paired-summary.json; new raw logs there. Oldcontrols review-evidence-old-corrected-20261113/logs, guardedrangeold review-rounds-old-guarded-20261113/logs. Newguardedrange reporttmp.94atIdyNIk; Coveragetmp.4rUEvV0bGT; reprotmp.geYCh80C9f; buildtmp.q7tcfFXrY0. No measured regression and no claim of improved pass rate or guaranteed compliance.
Earlier source candidates1/2 had guardedrange9/10; snapshots review-evidence-candidate1-guarded-20261113 and review-evidence-candidate2-20261113 retained. Third/final source candidate states incomplete reports retain ALL current input ranges including fix range, and third-round asks cover all ranges without offering another round. Cause of individual LLM choices is not proven. No fourth source fix or weakened rubric; exclusion fallback not needed after final40/40.
New rounds judge calibrated7/7 expected(2validPASS5invalidFAIL), tmp.Dvp6dVHBGD. New scaffold nonempty guard caught by package check; only affected range case remeasured10+10 under identical guarded fixtures. Existing three controls unchanged. Real-Git handoff property executes the skill assignment against advanced main in isolated HOME/TMPDIR and returns the branch point.
Wiki review-base-merge-base updated once; wiki check stale warnings only/no errors. No review/full/tagged-land gates yet. No push/attachments; all original approvals remain valid. After this lands close DBT0165.01-.03 then continue239-241 serially inline.
