---
parent: bugs/2026-11-07-flaky-eval-land-tidy-temp-worktree
depth: minimal
id: PLN-0243
created: "2026-11-13 11:31:16"
hash: ppdjuup
started: "2026-11-13 11:31:36"
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
- [ ] Red: add `TestLandGateWorktreeOwner`, then run `scripts/test ./internal/plugincheck -run 'TestLandGateWorktreeOwner|TestLandTidyPath'` and the strengthened case on old text to show that the owner rule is absent.
- [ ] Change: capture and use the absolute owner in tidy's temporary-worktree instructions, adjust only needed caps, update the existing wiki fact, run focused checks and formatting, then compare `land-tidy-temp-worktree` and `land-tidy-temp-refused` for at least 10 runs per arm and retain the evidence.
- [ ] Commit: fix(land): scope temporary gate commands to their owning repository

## State

### Findings

Task 1 controls pass: loaded graders reject missing/report-only/failed audits and forced-removal logs; actual Git wrapper audits distinguish correct owner, wrong repo, late removal and a different removed worktree. Focused plugincheck package passes after watched RED, diff formatting clean.
Identical corrected-fixture old-text baseline: land-tidy-temp-worktree 9/10, land-tidy-temp-refused 9/10 full-case passes. Raw logs and plugin snapshot /Users/iyay/.wwt/acta/gate-owner-old-20261113; matching old binary bin/acta. These are measurements, not passing land gates.
