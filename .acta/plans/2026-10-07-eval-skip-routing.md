---
parent: debt/2026-10-07-routing-eval-set
depth: minimal
closes: [SPC-0099]
id: PLN-0108
created: "2026-10-07 09:23:50"
hash: gxbjwes
started: "2026-10-07 09:25:33"
finished: "2026-10-07 09:34:39"
---
# Eval skips routing Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `scripts/eval` with no case filter runs every case except `routing-*`, and a run with `--tag` or `--case` runs exactly what was asked.

**Spec:** .acta/specs/2026-10-07-eval-skip-routing.md

**Tests:** `scripts/test ./internal/plugincheck/ ./internal/evalomp/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Never run a paid eval: no `scripts/eval` without `--bare` or a dry form that starts no model. Prove the `--case` form by reading `claude plugin eval --help` and `internal/evalomp` code, and by a test that runs the script with `claude` and `go` replaced by a stub on PATH that prints its arguments.
- `plugin/skills/land/SKILL.md` does not change. Keep every flag `TestEvalScriptFlags` already requires.
- Run tests with `scripts/test`, never bare `go test`. Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02

Task 01 and Task 02 touch different files and can run together.

### Task 01: default eval run skips routing cases

**Files:** `scripts/eval`, `internal/plugincheck/evals_test.go`, and `internal/evalomp/` only if `acta eval-omp` cannot take the case list the script passes (then the smallest change there, with its own test).

**verify:** For both the Claude path and the `--omp` path: with no `--tag` and no `--case`, the command that starts the eval names every case folder under `plugin/evals/` except those starting with `routing-`, and no `routing-*` case; with `--tag routing` or any `--case`, the user's arguments pass through and no case list is added. List each path and argument shape the test ran (no args, `--tag routing`, `--case x`, `--omp` with none, `--omp --case x`).

- [x] Red: in `internal/plugincheck/evals_test.go`, a test runs `scripts/eval` with a temp PATH whose `claude` and `go` stubs print their arguments and exit 0, and asserts the case lists above for each argument shape; run `scripts/test ./internal/plugincheck/ -run <TestName>` and watch it fail because the default run passes no case filter.
- [x] Green: in `scripts/eval`, when the arguments hold no `--tag` and no `--case`, build the list of folders in `plugin/evals/` that hold `prompt.md` or `case.yaml` and do not start with `routing-`, and pass them with `--case` in the form `claude plugin eval --help` and `acta eval-omp` accept; add a plain comment saying the routing set is a baseline measure and in the land gate would add 54 runs to every land.
- [x] Commit: `fix(eval): default run skips the routing baseline cases (SPC-0099)`

### Task 02: Version 0.1.31

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.31, and `internal/plugincheck` passes. List each file and the version it holds.

- [x] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [x] Green: change 0.1.30 to 0.1.31 in all three files.
- [x] Commit: `chore: version 0.1.31`

## Fix round 1

### Task 03: Fix round 1, routing cases move to their own eval dir

**Files:** `scripts/eval`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/routing_evals_test.go`, `plugin/evals/routing-*` moved with `git mv` to `plugin/evals-routing/routing-*`, `internal/cli/eval_omp.go` and its test, `plugin/evals/FACTS.md` (one short paragraph on the new dir).

**verify:** A default `scripts/eval` run (Claude and `--omp`) sees no routing case and every other case, because `plugin/evals/` holds no `routing-*` folder; `scripts/eval --eval-dir evals-routing` and `scripts/eval --omp --eval-dir evals-routing` reach the 18 routing cases; `scripts/eval` runs under macOS `/bin/bash` 3.2 with any arguments. List each path and argument shape tested, and the bash used.

- [ ] Red: tests that fail today: `plugin/evals/` holds a `routing-*` folder; `TestRoutingEvalCases` and the scaffold refusal test read `plugin/evals-routing/`; `acta eval-omp --eval-dir evals-routing --case x <plugin>` loads cases from that dir; `scripts/eval` run with `/bin/bash` and stub `claude`/`go` on a temp PATH (HOME and TMPDIR pinned to temp dirs) passes `--eval-dir evals-routing` through untouched and adds no `--case`.
- [ ] Green: revert the case-list logic in `scripts/eval` from commits cc8d985 and b627f63 (the script goes back to passing `"$@"` only) and replace their test with the tests above; `git mv` the 18 folders; add `--eval-dir` (default `evals`) to `acta eval-omp`; add a plain comment in `scripts/eval` that the routing baseline lives in `plugin/evals-routing/` so the land gate does not run it, and is run with `--eval-dir evals-routing`.
- [ ] Commit history: fold the formatting-only commit 2a9b818 into cc8d985 with a fixup and `git rebase -i`-free autosquash (`GIT_SEQUENCE_EDITOR=: git rebase --autosquash`), own unpushed commits only, before the fix commit.
- [ ] Commit: `fix(eval): routing baseline lives in its own eval dir (SPC-0099 fix round 1)`
