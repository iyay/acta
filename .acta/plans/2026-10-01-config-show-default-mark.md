---
parent: specs/2026-10-01-config-show-default-mark-design
depth: minimal
id: PLN-0067
created: "2026-10-01 06:46:37"
hash: pfi9frt
started: "2026-10-01 06:50:22"
finished: "2026-10-01 06:51:50"
---
# Config Show Default Mark Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta config show` prints `plan_depth: full (default)` when no file sets it, and setup asks for any value marked `(default)`.

**Spec:** `.acta/specs/2026-10-01-config-show-default-mark-design.md`

**Tests:** fast `scripts/test ./internal/cli ./internal/plugincheck`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The mark is exactly ` (default)`, after the value, on the text output only. JSON keeps `"plan_depth": "full"`.
- Comments are plain English a 10-year-old can read. They say why, not what.

## Waves

- Wave 0: Task 0 (added at build start: the baseline was red).
- Wave 1: Task 1, Task 2 (no shared files).

### Task 0: Keep the executor test away from the repo's .acta.yaml

**Files:**
- Modify: `internal/cli/cli_test.go`

**verify:** No config show test reads the real repo's `.acta.yaml`. List every config show test that runs without its own cwd and what cwd each gets.

- [x] Failing test: `TestVoiceSetExecutorKeepsOtherFields` already fails on main, because it runs in the package folder and show reads `build_executor: dispatch` from the repo's `.acta.yaml`.
- [x] Code: the test calls `t.Chdir(t.TempDir())` first, so show finds no repo file.
- [x] Commit: `executor test runs outside the repo so .acta.yaml cannot leak in`

### Task 1: Mark an unset plan_depth in config show

**Files:**
- Modify: `internal/cli/config_cmd.go`
- Test: `internal/cli/config_repo_test.go`

**verify:** The ` (default)` mark shows only when neither the user file nor `.acta.yaml` sets plan_depth. List every source checked (none, global full, global minimal, repo full, repo minimal) and the line each prints; JSON output never carries the mark.

- [x] Failing test: the "nothing set anywhere" test expects `plan_depth: full (default)` and a new test with global `plan_depth: full` expects no mark; both fail because show prints bare `full` for both.
- [x] Code: in the show branch, when `v.PlanDepth` is empty, add ` (default)` to the text line only.
- [x] Commit: `config show marks plan_depth when it is only the default`

### Task 2: Setup asks for values marked (default)

**Files:**
- Modify: `plugin/skills/setup/SKILL.md`

**verify:** No setup path treats a `(default)` value as already set. List every place the skill decides whether to ask.

- [x] Failing test: none needed for skill prose; `scripts/test ./internal/plugincheck` must stay green after the edit.
- [x] Code: First run step 2 adds that a value marked `(default)` is not set yet, so setup asks for it.
- [x] Commit: `setup asks for config values marked (default)`
