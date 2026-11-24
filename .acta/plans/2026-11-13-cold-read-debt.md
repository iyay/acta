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

- [x] Contradiction grader negation guard misses plurals: "No conflicts: 30 days and 7 days." passes; add s? and a must-fail sample.
- [x] Commit: `polish: review notes for PLN-0241`

## Review notes

- Review depth auto picked standard (skill text changed); round 1 CLEAN on both axes; polish 8324cbf reviewed by one reviewer, CLEAN. BLOCKER 0, NOTE 10, pre-existing 0. Notes: guard trade-offs on negations elsewhere in a reply, false-claim gain is a grader change, shape folder has no spare line, scaffold comment style.

## State

### Next

Round 1 CLEAN on both axes (0 BLOCKER). Polish 8324cbf is under its review. Then: debt for relay, review-notes line, acta:land with full suite and eval gate --tag=shape. No push.

### Findings

Quota blocker of 2026-11-14 cleared on 2026-11-16; all new-arm runs scored with zero quota errors.
Final 10-run arms, same graders (logs in /Users/iyay/.wwt/acta/cold-read-debt-evidence-20261114/logs): relay old 6 new 7 and 8; read-evidence old 9 new 10; contradiction old 9 new 10 (widegrader2); false-claim old 0 new 10 (grader change, not a skill effect); accept old 9 new 10.
Contradiction regex is phrase-brittle; widened in 085242b from three real replies; plural negation guard fixed in polish 8324cbf.
Relay: about 1 to 2 of 10 runs still summarize A1-A6 without quotes; one haiku judge false negative seen on a full 11-quote table.
Old-arm runs from task2/old need --trust-plugin when no terminal is attached.
