---
parent: bugs/2026-11-07-flaky-eval-land-tidy-temp-worktree
depth: minimal
id: PLN-0243
created: "2026-11-13 11:31:16"
hash: ppdjuup
started: "2026-11-13 11:31:36"
finished: "2026-11-13 11:40:49"
---
# Run temporary gate operations through their owning repository

**Goal:** A tidy land removes its clean temporary gate worktree from the owning repository before moving the parent, regardless of the shell's current repository.
**Spec:** `.acta/specs/2026-11-13-gate-worktree-owner.md` (SPC-0210)
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched, or `scripts/eval --case <name>` for one case alone; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build inline in a new worktree under `~/.wwt/acta/`; no implementer subagents, new CLI, setting, force, recursive deletion, guard bypass or push.
- Keep the owning repository and tested temporary worktree as separate absolute paths. Preserve red-gate stops, genuine refusal handling, existing cleanup permissions and final plan checks. BUG-0092 remains outside this plan.
- Run scaffold and mutation probes in isolated copies with fresh HOME and TMPDIR. Use existing evalomp loader/grader interfaces and wrapper patterns, not a new framework. Never run Git in retained eval sandboxes.
- Keep the successful-remove-before-parent grader. Validate each changed fixed grader with passing and failing samples and actual wrapper/filesystem controls before scoring the instructions.
- After Task 1, capture unchanged skill text and its matching binary in an ordinary isolated copy outside system temp. Compare each named case separately with `scripts/eval --case <name> --runs 10 -j 8 --keep-temp`, old versus new under identical corrected fixtures and runner settings, with the matching arm binary on PATH.
- Retain raw counts and traces. Report remaining failures and any regression; never weaken a grader or repeat an unchanged arm to chase green. Full suites wait for land. Raise only necessary text caps.

## Build brief

- Land this correction on main first. Then resume PLN-0242 in its existing `gate-output-cleanup` worktree, preserving its approved cleanup policy and this owner-repository rule where they overlap, and run the required combined-tree gates. Continue PLN-0237 through PLN-0241 afterward; their approvals remain valid.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Observe removal in the owning repository before parent movement

**Files:** `plugin/evals/land-tidy-temp-worktree/{prompt.md,scaffold.sh,graders/temp-removed.md}`, new `plugin/evals/land-tidy-temp-worktree/graders/owner-removal.md`, new `internal/plugincheck/gate_worktree_owner_test.go`
**Seam:** Real eval scaffolding and Git wrappers, loaded through `evalomp.LoadCases` and graded through `evalomp.Grade` over actual registration/effect logs.
**verify:** A distinct outer repository cannot remove the owner's registered temporary gate worktree. Accept the matching worktree's successful owner-scoped removal before parent movement; reject wrong-repository attempts treated as success, a different worktree's removal, cleanup only after parent movement and force. Audit actual registration, exit statuses and order independently of the final report. Keep the existing removal-order protection and list each positive/negative control exercised.
- [x] Red: add `TestGateWorktreeOwnerGraders` and `TestGateWorktreeOwnerAudit`, then run `scripts/test ./internal/plugincheck -run 'TestGateWorktreeOwner|TestEvalCases|TestEvalGraderPatternsCompileInOmp'` to expose missing ownership/effect evidence and invalid stand-ins before changing the fixture.
- [x] Change: strengthen the existing case with a distinct outer repository and owner-aware registration/effect logging, add the fixed ownership grader, validate positive/negative samples and actual wrapper controls, then capture the old arm and run both named comparison cases on unchanged instructions.
- [x] Commit: test(land): verify temporary gate cleanup belongs to its repository

### Task 2: Pin temporary gate commands to their owning repository

**Files:** `plugin/skills/land/tidy.md`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/budget_test.go`, `.acta/wiki/land-tidy-history.md`
**Seam:** Land's standalone tidy instructions, checked through focused plugincheck contracts and Task 1's behavioral case.
**verify:** Every temporary-worktree creation, listing and removal instruction uses the captured absolute owning repository rather than an ambient shell directory; the tested worktree path remains distinct. A clean removable worktree is confirmed removed before parent movement. Wrong-repository errors never count as cleanup success or as proof of a file refusal. Existing red-gate and genuine-refusal policies retain their meaning; Task 1 owns the adversarial command-order cases.
- [x] Red: add `TestLandGateWorktreeOwner`, then run `scripts/test ./internal/plugincheck -run 'TestLandGateWorktreeOwner|TestLandTidyPath'` and the strengthened case on old text to show that the owner rule is absent.
- [x] Change: capture and use the absolute owner in tidy's temporary-worktree instructions, adjust only needed caps, update the existing wiki fact, run focused checks and formatting, then compare `land-tidy-temp-worktree` and `land-tidy-temp-refused` for at least 10 runs per arm and retain the evidence.
- [x] Commit: fix(land): scope temporary gate commands to their owning repository

## Review notes

- Round 1: auto selected standard, full tier, because agent instructions and a delete path changed; delete-path floor and deep lens applied. Spec and Standards CLEAN; BLOCKER 0, NOTE 0, pre-existing 2 (known residual skipped gate under BUG-0072 and report-policy failures under BUG-0080, already tracked; no new bug). Narrow plugincheck checks pass; isolated production reversion makes the owner contract test fail with fresh HOME/TMPDIR. No polish. Comparative counts remain target9/10 versus9/10, refusal9/10 versus8/10; no aggregate improvement or eliminated-flakiness claim. No reviewer worktrees or undispatched commits.

## State

### Next

Land with tidy after the recorded green gates, then resume PLN-0242 in its existing worktree, preserving both approved policies and running combined-tree gates. Continue PLN-0237 through PLN-0241 afterward, without repeating approvals.
Do not claim historical flakiness eliminated; raw comparison counts remain mixed. WhatsApp text only unless documents requested.

### Findings

Review auto/standard/full CLEAN, no new findings; the two known residual behaviors are tracked by BUG-0072 and BUG-0080. Owner-repository lesson is in land-tidy-history wiki; no additional debt from this review.
Full suite PASS: 28 package results, no FLAKY lines. Tested commit071e1a3e6e4548d68da2c310cd48937347fa588a, treecd515f93cbf266f0830777101d3eca90942df92f; .git/acta-baseline.md retains full log.
Land-tag eval12/13 initially; sole red land-tidy-temp-refused passed its one permitted known-flake retry and was recorded on BUG-0080. Target land-tidy-temp-worktree passes both ownership and order graders. Whole hook gate is not required: no plugin/hooks or plugin/references changes.
Gate log .git/acta-owner-eval-20261113T044402Z-34868.log; retry .git/acta-owner-retry-refused-20261113T044636Z-76076.log.
Old/new10-run counts: target9/10 versus9/10, refusal9/10 versus8/10; refusal fixed-effect checks10/10 both. Summary and raw evidence /Users/iyay/.wwt/acta/gate-owner-old-20261113/logs/summary.json. No aggregate improvement claim.
Old failuree-NWbBoP wrong repo then late removal; new failuree-7r9qul skips detached gate entirely. Refusal judge issues remain under the separate approved PLN-0242. Do not expand this fix's scope or discard its worktree.
No reviewer worktrees, source changes or undispatched commits after review; only planning records changed after the tested tree.
