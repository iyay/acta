---
parent: debt/2026-11-13-cold-read-evidence
depth: minimal
closes: [DBT-0168.01, DBT-0168.02, DBT-0168.03, DBT-0168.04]
after: [PLN-0240]
id: PLN-0241
created: "2026-11-13 06:13:11"
hash: qeb41h1
started: "2026-11-14 06:29:59"
finished: "2026-11-16 11:44:52"
---
# Cold read evidence survives grading and relay

**Goal:** Cold-read cases measure defect recognition and complete relay, while Read evidence has a checkable format.

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

- Land order: start only after PLN-0240 lands on main; these serial plans share plugincheck and eval files.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Remove prompt hints and validate grader meanings

**Files:** `plugin/evals/cold-read-planted-contradiction/{prompt.md,graders/names-both-lines.md}`, `plugin/evals/cold-read-false-claim/prompt.md`, `plugin/evals/polish-full-review/graders/{reviewer-dispatched.md,standards-reviewer-dispatched.md}`, `plugin/evals/review-auto-signal/graders/{spec-reviewer-dispatched.md,standards-reviewer-dispatched.md}`, `plugin/evals/review-depth-floor/graders/{spec-reviewer-dispatched.md,standards-reviewer-dispatched.md}`, `plugin/evals/review-clean-has-evidence/graders/brief-asks-coverage.md`, `internal/plugincheck/workflow_debt_evals_test.go`, `internal/plugincheck/evals_test.go`
**Seam:** loaded contradiction grader against sample final replies and static checks of prompt leakage and all seven matcher comments.
**verify:** Quoting two durations without recognizing conflict fails, while explicit contradictions with spaced or hyphenated durations pass in either order. Neither planted-defect prompt tells the run to relay every finding. All seven specified comments describe the current regex matcher; none still claims literal substring matching, including files whose grader type an earlier plan changed. Do not change a grader's matching behavior merely to fix its comment.
- [x] Red: add `TestWorkflowDebtColdReadGraders` samples and prompt/comment checks, then run `scripts/test ./internal/plugincheck -run 'TestWorkflowDebtColdReadGraders|TestEvalGraderPatternsCompileInOmp'` to expose false positives, valid hyphenated failures and stale descriptions.
- [x] Change: require conflict recognition with a compatible JavaScript regex, remove the relay hints, correct all seven comments to match their final grader types, and run both planted-defect cases for 10 old-text runs under these corrected graders.
- [x] Commit: test(shape): measure contradictions without relay hints

### Task 2: Require structured reading evidence and preserve every finding

**Files:** `plugin/skills/shape/cold-read.md`, `internal/plugincheck/skill_shape_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`, new `plugin/evals/cold-read-findings-relay/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`, new `plugin/evals/cold-read-read-evidence/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`, `plugin/evals/cold-read-false-claim/graders/brief-asks-what-was-read.md`
**Seam:** cold-read Resolve and reviewer report contract, with planted completed reviewer reports.
**verify:** Every returned finding reaches the user individually with its spec quote; a six-finding report loses none, and a capped five-finding report preserves its cut-count line. The returned report must have `### Read`, nonempty spec-section entries and reference paths or explicit none; missing, empty or generic “all of it” evidence triggers the existing incomplete-report choice and waits. Valid evidence proceeds only to final spec approval, never automatic edits or planning. The eval prompt does not state the behavior being tested.
- [x] Red: add `TestColdReadStructuredEvidence` and the relay/evidence cases with valid, missing and empty reports, then run `scripts/test ./internal/plugincheck -run 'TestColdReadStructuredEvidence|TestShapeColdRead|TestEvalCases|TestEvalLLM'` and the new cases on old text.
- [x] Change: define the reading-evidence form and incomplete-report checks, keep the existing no-drop rule, align required strings and narrow judge entries, then compare `cold-read-findings-relay`, `cold-read-read-evidence`, `cold-read-planted-contradiction`, `cold-read-false-claim` and `cold-read-accept` for 10 runs per arm.
- [x] Commit: fix(shape): require structured Read evidence and relay every finding

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [ ] Contradiction grader negation guard misses plurals: "No conflicts: 30 days and 7 days." passes; add s? and a must-fail sample.
- [ ] Commit: `polish: review notes for PLN-0241`

## State

### Next

Both tasks done. Next: wiki check, acta:review over merge-base main..HEAD, then acta:land with full suite and eval gate (--tag=shape). No push.

### Findings

Prior0240land0bd5521/DBT0167closure0cdd165, originalprogress10/14in4/5groups. Task1code19fa027done/readback/wave1; compatiblefixedgrader testsRED thenGREEN0.032s/package14.817s. Finaloldskill contradiction10/10 tmp.zdXwjZESZd, falseclaim10/10 tmp.6oMKZDv7yr underoldbrief. Original7/10+8/10wereprovedgraderfalse-negatives, retainedseparately, no skill-effectclaim.
Task2uncommittedcandidate: intendedstructuredReadRED0.065s; loadedbriefgraderRED0.037s->GREEN0.033s. NewLLMcalibration initial8/9expected (droppedA6falsePASS tmp.08inJbGGdL); explicit11quoteaudit/reference-list-notreplydistinction fixes it, final9/9expected tmp.eNMQBZ1YE4. BothscaffoldshandrunfreshHOME/TMPDIR; committedspec/reportpremisestrue. Source requires###Read,namednonemptySpecsections,Referencepaths-or-explicitnone; missing/empty/genericreportexistingincompletechoice/wait. Clarifiesfivecapdoesnotlimitrelay, preservesindividualquotes/cutcount. FocusedGREEN0.077s/package13.784s; gofmt/diffcheckclean, all5old/newgrader/fixturehashchecksTrue. Shape179lines/cold-read3516bytes, no safetycuts.
Oldcorrectednativearms: relay6/10 tmp.ItlS8Pm6Bu (actuale-R5aaFcsummarizes11findingswithoutquotes), Read9/10 tmp.4GKUXB7QYE (actuale-EOdI3yacceptsgenericall-of-it), falseclaim0/10 tmp.DhrT6wzZqK (oldbriefdoesnotrequestnewfields; other2graders10/10), accept9/10 tmp.oNmp5q16lC. ReuseTask1contradiction10/10armbecauseidenticalskill/fixture/grader; hashesconfirmed. Matchingprivateoldclone task2/old, binoldandbin-new under cold-read-debt-evidence-20261114.
BLOCKER externalquota2026-11-14 11:20: all50new-armrunsungraded,44agentAPIerrors+6judgethrows: You've hit your weekly limit · resets 10pm (Asia/Jakarta). Reports relaytmp.zewAoAqnZM, Readtmp.FwuoepALop, contradictiontmp.aWglQf994m, falseclaimtmp.rlilj07owV, accepttmp.EYrm0GilWo. quota-summary.json capturesexactcounts/logs. These0scoresarenot skillregression/ordinaryflakes; do nottickTask2done, commitimplementation, review, fullsuiteorland. Do notretryquota unchangedbefore advertised22:00WIB reset, switchaccounts/backendsorweakengates. No sourcefailurefromquota.
