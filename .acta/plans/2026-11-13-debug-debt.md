---
parent: debt/2026-11-13-debug-evidence
depth: minimal
closes: [DBT-0167.01, DBT-0167.02, DBT-0167.03]
after: [PLN-0239]
id: PLN-0240
created: "2026-11-13 06:13:10"
hash: wjfpdle
---
# Debug uses private probes of the right checkout

**Goal:** Debug probes copy the checkout being investigated and keep every pre-proof write within private probe storage.

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

- Land order: start only after PLN-0239 lands on main; these serial plans share plugincheck and eval files.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Probe the checkout under investigation

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`, new `plugin/evals/debug-worktree-source/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`, `.acta/wiki/skill-escape-lines-skip-proof.md`
**Seam:** debug control instructions and the `debug-worktree-source` eval with different main/worktree contents.
**verify:** Every copied control uses the absolute root of the tree under investigation, captured before changing directories; invocation from main and later shell resets cannot silently change it. A main-only copy cannot prove a worktree defect. The evidence names the tested checkout and shows the symptom there plus a control with the suspected cause removed. No $OLDPWD source remains in the control recipe.
- [ ] Red: replace the old required-copy assertion with `TestDebugControlSource` and add the differing-checkout case, then run `scripts/test ./internal/plugincheck -run 'TestDebugControlSource|TestDebugProof|TestEvalCases'` and the new case against old text.
- [ ] Change: capture and use an explicit absolute source root throughout the control recipe, update focused rules and wiki facts, then compare `debug-worktree-source` and `debug-independent-check` for 10 runs per arm.
- [ ] Commit: fix(debug): copy controls from the checkout under investigation

### Task 2: Cover every pre-proof write with private probe storage

**Files:** `plugin/skills/debug/SKILL.md`, `plugin/skills/debug/root-cause-tracing.md`, `plugin/skills/debug/race.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`, `plugin/evals/debug-independent-check/graders/no-repo-write.md`, new `plugin/evals/debug-probe-isolation/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`
**Seam:** debug write-boundary instructions and trace/file evidence from `debug-probe-isolation`.
**verify:** Tests with caches, polluter hunts, profiler outputs and file-tool writes obey the same pre-proof boundary; only explicit metadata exceptions permit repo planning-file writes, never source edits. Probe storage is newly created with restrictive permissions and exclusive creation: an unset TMPDIR, pre-existing name, symlink, failed creation or shell reset cannot cause writes in the source tree or an old/shared folder. Each shell write checks its folder in the same command; file-tool writes name verified absolute probe paths. The existing no-repo-write grader no longer excuses cache writes in the source checkout.
- [ ] Red: add `TestDebugProbeIsolation` and planted write/collision scenarios, then run `scripts/test ./internal/plugincheck -run 'TestDebugProbeIsolation|TestDebugProbeFolder|TestDebugRerunsKnownFlake|TestEvalCases'` and the new case on current text to expose unguarded outputs and shared probe reuse.
- [ ] Change: use exclusive private creation supported by the eval sandbox, carry checked paths through all referenced recipes and file tools, state bug/flake metadata exceptions precisely, update the isolation grader, then compare `debug-probe-isolation`, `debug-independent-check`, `debug-order-dependent` and `debug-performance` for 10 runs per arm.
- [ ] Commit: fix(debug): isolate all pre-proof writes in private probe folders

