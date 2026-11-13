---
parent: bugs/2026-11-10-flaky-eval-land-tidy-temp-refused
depth: minimal
id: PLN-0242
created: "2026-11-13 07:45:44"
hash: entlb0i
---
# Approved cleanup of temporary gate outputs

**Goal:** An agent can remove proven gate outputs from its own temporary worktree after typed approval for that exact file list, while every unsafe or unapproved batch stays untouched.
**Spec:** `.acta/specs/2026-11-13-gate-output-cleanup.md` (SPC-0209)
**Tests:** fast `scripts/test ./internal/plugincheck ./internal/hook`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build inline in a new worktree under `~/.wwt/acta/`; no new CLI, setting or change to `internal/hook/destructive.go`; never bypass a tool rejection, force, recursively delete or push.
- The exception covers only session-created temporary gate worktrees and one displayed list per explicit typed approval; main and plan/build worktrees retain their policy.
- Record inventory including ignored files before the gate; record completed producer evidence, regular-file identities and content fingerprints afterward in private storage outside the worktree; unknown proof never authorizes cleanup.
- Check the entire approved batch before the first deletion; any invalid member cancels all deletion, and changed evidence needs a fresh list and approval.
- Each eval scaffold guards an empty directory; manual scaffold, grader and mutation probes use fresh HOME and TMPDIR outside real checkouts. Reuse the existing runner, grader types and fixture patterns, not a new test framework.
- Fixed grader samples call `evalomp.LoadCases` and `evalomp.Grade`; trace-based judge rubrics are checked against planted valid and invalid completed runs with the existing evaluator, never a fake judge that merely echoes the expected vote.
- Capture the old plugin and a binary embedding the old core after Task 1 and before Task 2; build the new binary separately after Task 2 and use the correct arm's binary on PATH. Both arms use identical corrected graders/scaffolds, model and runner settings. Keep comparisons in isolated copies outside the system temp tree, following the eval wiki.
- Compare all three new cases and both neighboring cases for at least 10 runs per arm, one `scripts/eval --case <name> --runs 10 -j 8 --keep-temp` call per case. Keep raw evidence; improve the intended permission behavior without regressing unchanged boundaries or neighboring behavior.
- Raise required text caps rather than trim safety instructions. Full suites run only at land. Do not import the unlanded PLN-0237 fixture.

## Build brief

- Land on main before resuming PLN-0237. Keep its reviewed `land-debt` branch intact; resolve overlapping changes by preserving both approved behaviors, then run the gates required for that combined tree. Continue the four already-approved debt plans afterward.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Measure permission and file-boundary behavior from real effects

**Files:** `plugin/evals/land-tidy-temp-refused/{prompt.md,scaffold.sh,graders/two-choices.md}`, new `plugin/evals/land-gate-output-approved/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md,graders/effects.md}`, new `plugin/evals/land-gate-output-no-approval/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md,graders/effects.md}`, new `plugin/evals/land-gate-output-boundaries/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md,graders/effects.md}`, new `internal/plugincheck/gate_output_evals_test.go`, `internal/plugincheck/evals_test.go`
**Seam:** Loaded real eval graders over planted tool traces and filesystem audits, plus the three new cases and the existing refused-removal case.
**verify:** A directly typed approval for a displayed exact list permits only its unchanged proven outputs; no approval, refusal, emoji, blanket session approval, forwarded/reference, peer, relayed or different-list approval deletes nothing. Separately planted tracked, pre-existing, unknown-producer, symlink/ancestor-symlink, traversal/outside-root, changed-content, replaced-file, changed-worktree and missing-inventory cases preserve every file; a mixed safe/unsafe batch preserves even its safe member. Each boundary has its own controlled candidate and literal expected result. Use a fixed effects grader over an external filesystem audit, backed by a trace-focused permission/procedure judge. Actual post-run file contents/identities and worktree state must back the verdict; quoting checks or an inaccurate report cannot pass. A green land with pending cleanup still continues and reports what remains. Run the existing `land-tidy-temp-worktree` as an unchanged neighbor.
- [ ] Red: add grader samples and controlled permission/boundary scenarios, then run `scripts/test ./internal/plugincheck -run 'TestGateOutputEval|TestEvalCases|TestEvalLLMGradersAreOnlyTheJudgementCalls|TestEvalGraderPatternsCompileInOmp'` and planted judge samples to expose the legacy last-message-only permission contract before changing its graders.
- [ ] Change: implement the three focused cases with direct-authority prompts, completed producer evidence and external filesystem audits, align the old refused-removal grader to the approved policy without weakening unauthorized-deletion controls, update the narrow judge allowlist and validate positive/negative samples, then capture the unchanged policy and old binary for Task 2 comparisons.
- [ ] Commit: test(land): measure approved gate cleanup and unsafe batch preservation

### Task 2: Apply the narrow cleanup permission across core and land

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`, `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/budget_test.go`, `.acta/wiki/land-tidy-history.md`, `.acta/wiki/eval-red-case-debug.md`
**Seam:** `hook.SessionStart` and its generated fallback parity, land/tidy's workflow, and the real cleanup cases from Task 1.
**verify:** Every emitted core/fallback and land instruction agrees on the limited exception. Inventory/provenance, typed-list approval and whole-batch revalidation precede any deletion. Only individual approved outputs are removed; no recursive/forced deletion or tool-rejection bypass is introduced, and all existing destructive-command guards remain effective. Invalid/missing proof, stale identity/content or absent permission preserves the batch; a plain remove refusal preserves the remaining worktree. Main/build/other-session trees cannot use this exception. A green approved merge does not wait for optional cleanup permission, a red gate still stops, and the final report accurately distinguishes deleted outputs from retained worktrees. List every instruction surface checked; Task 1 owns the individual adversarial cases.
- [ ] Red: add the focused SessionStart/fallback and land-contract regressions, run `scripts/test ./internal/hook -run 'TestSessionStartGateOutputCleanup|TestDefaultRulesFile'` and `scripts/test ./internal/plugincheck -run 'TestLandGateOutputCleanup'`, then measure Task 1 cases on the old arm to show that approved cleanup is refused or the report follows the superseded policy.
- [ ] Change: add the bounded exception and complete procedure without changing the destructive guard, regenerate fallback with `scripts/test ./internal/hook -run TestDefaultRulesFile -args -update`, adjust only necessary caps/expectations, update wiki facts, run focused hook/plugincheck tests and the formatter, then measure the same five cases on the new arm and preserve raw counts and traces.
- [ ] Commit: feat(land): clean approved outputs from temporary gate worktrees

