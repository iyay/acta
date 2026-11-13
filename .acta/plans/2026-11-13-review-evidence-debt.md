---
parent: debt/2026-11-13-review-evidence
depth: minimal
closes: [DBT-0165.01, DBT-0165.02, DBT-0165.03]
after: [PLN-0237]
id: PLN-0238
created: "2026-11-13 06:13:07"
hash: f3qekfa
started: "2026-11-13 18:44:11"
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
- [ ] Red: add focused `TestReviewIncompleteRanges` and `TestBuildReviewMergeBase` checks plus the incomplete-rounds case, then run `scripts/test ./internal/plugincheck -run 'TestReviewIncomplete|TestBuildReviewMergeBase|TestReviewVerdictNeedsEvidence|TestEvalCases'` and the case on old text to expose the lost ranges and contradictory CLEAN path.
- [ ] Change: state the cumulative incomplete-range rule and round-3 stop, fix build's range handoff, update the wiki, then run the focused checks and measure `review-incomplete-rounds`, `review-clean-has-evidence`, `review-false-blocker` and `build-close-no-suite` for 10 runs per arm.
- [ ] Commit: fix(review): preserve incomplete ranges within the round budget

## State

### Next

Task1 DONE a0d56f9; ticks/readback3/3 and wave1 committed. Corrected old controls: Coverage10/10,repro10/10,buildclose10/10 in review-evidence-old-corrected-20261113/logs. Guarded incomplete-range old10/10 in review-rounds-old-guarded-20261113. All final fixtures match between arms.
Task2 ACTIVE and uncommitted. RED/green recorded for cumulative range wiring and a real-Git moved-parent handoff test. Final focused checks PASS; package13.799s PASS before latest review-only prose. Build cap11094/review11658. No full suite yet. New rounds judge calibration7/7 expected, tmp.Dvp6dVHBGD.
Candidate1 guarded results39/40: target9/10 and three controls10/10; e-Ru1b7q offered another review in third-round ask. Candidate2 also39/40: e-I7KrQ9 table has correct inputs, but ask mentions only original range and drops fix range. Interpretive cause not proven; source ambiguity is a hypothesis. Both snapshots/raw results retained, no rubric weakened.
Candidate3 explicitly says an incomplete report leaves ALL current input ranges incomplete, including fix range; third-round ask concerns all unreviewed ranges and never offers another round. This is the THIRD and last source candidate, not an unchanged retry. Matching binary review-evidence-candidate3-bin-20261113/acta; measurements sessions57343(target,Coverage) and91752(repro,buildclose), logs/logs/new-*.log,10 per case. Await all full-case results.
If candidate3 regresses, follow approved global fallback: exclude regressing review text, retain DBT0165.01 open and remove it from closes; keep nonregressing build handoff and calibrated evidence fixes for .02/.03. Do not attempt a fourth source fix. Record exact narrowed scope through acta state/rulings before review; no fresh permission needed for the already-approved exclusion rule.
If no regression, commit Task2 with Task:PLN-f3qekfa#2, tick/readback/wave2commit; update existing review-base-merge-base wiki once, review two fresh read-only axes, then full suite and tagsreview/build. Do not push or attach docs. Continue approved239-241 after238 lands; do not redo completed237/244.

### Findings

Task1 committed a0d56f9, tick/readback3/3 andwave1committed. Corrected old baselines Coverage10/10,repro10/10, buildsnapshotreview-evidence-old-corrected-20261113; native Coverage judge9/9 plantedexpected after proven delivery-count correction, reports/raw retained. Originalprecorrection9/10 runs excluded, not instructioneffects.
Task2 watchedRED for missing cumulative ranges and missing merge-base handoff. New real-Git property runs skillassignment underfreshHOME/TMPDIR with advancedmain; GREEN returnsbranchpoint. Initialcandidate1 focused0.141sPASS. Packagecaughtnewscaffoldmissingnonemptyguard plusobsoletewordingassertions; fixedsetup/updatedassertionstosame strongercontract, package13.799sPASS. No fullsuite yet.
New roundsjudge7/7 plantedexpected(2validPASS5invalidFAIL), tmp.Dvp6dVHBGD, private review-rounds-probes-20261113. Guarded oldrange case10/10 tmp.ocoebg8zFk. Beforeguard, old/newrange10/10each retained; onlyguardedpaircountsforfinalcase.
Candidate1 othercases all10/10: Coverage tmp.Y0PpCbA814, repro tmp.dI9ZNIA2XS, buildclose tmp.2PFK3zbYnp. Guardednewrange9/10 tmp.LlbHr0TzqJ: e-Ru1b7q retainedcorrectranges/BLOCKED butofferedanotherreview inRound3's askmenu. Do notclaimnonregression; firstcandidateactual39/40 vsold40/40. Snapshotreview-evidence-candidate1-guarded-20261113 andoriginalpreGuard snapshotretained.
Candidate2 clarifies that Round3 incomplete stops/asks aboutunreviewedwork withoutofferinganotherreview; subsequentland/fix/revert BLOCKERchoices applyonlytocomplete reports. CauseofLLMchoice notproven; ambiguitybetweenincompletebranch andBLOCKERmenuis hypothesis. WatchednativecaseRED precedesclarification; narrow0.031sPASS. Exactreviewcap11560, build11094. No grader weakened orunchanged rerun tochasegreen.
Candidate2 privatebinarybuilding session102? actualexecsession inturn; directoryreview-evidence-candidate2-bin-20261113/logs. Next runfournewcases10each onecasepercall withthismatchingbinary, compareimmutablecorrectedold40/40. CommitTask2onlyafteranalyzingresults; thenwikiupdateonce/review/full/taggedland. No reapproval/push/attachments.
