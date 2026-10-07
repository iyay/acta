---
parent: debt/2026-10-07-routing-eval-set
depth: minimal
closes: [SPC-0099]
id: PLN-0108
created: "2026-10-07 09:23:50"
hash: gxbjwes
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

- [ ] Red: in `internal/plugincheck/evals_test.go`, a test runs `scripts/eval` with a temp PATH whose `claude` and `go` stubs print their arguments and exit 0, and asserts the case lists above for each argument shape; run `scripts/test ./internal/plugincheck/ -run <TestName>` and watch it fail because the default run passes no case filter.
- [ ] Green: in `scripts/eval`, when the arguments hold no `--tag` and no `--case`, build the list of folders in `plugin/evals/` that hold `prompt.md` or `case.yaml` and do not start with `routing-`, and pass them with `--case` in the form `claude plugin eval --help` and `acta eval-omp` accept; add a plain comment saying the routing set is a baseline measure and in the land gate would add 54 runs to every land.
- [ ] Commit: `fix(eval): default run skips the routing baseline cases (SPC-0099)`

### Task 02: Version 0.1.31

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.31, and `internal/plugincheck` passes. List each file and the version it holds.

- [ ] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [ ] Green: change 0.1.30 to 0.1.31 in all three files.
- [ ] Commit: `chore: version 0.1.31`
