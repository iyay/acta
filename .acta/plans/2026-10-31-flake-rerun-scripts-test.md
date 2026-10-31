---
parent: specs/2026-10-31-flake-tracker-design
depth: minimal
id: PLN-0205
created: "2026-10-31 17:59:57"
hash: fr788tx
---
# Flake rerun in scripts/test implementation plan

**Goal:** Make `scripts/test --full` rerun each failed Go test alone up to 2 times and print a `FLAKY:` line for each one that passes on a rerun.
**Spec:** `.acta/specs/2026-10-31-flake-tracker-design.md` (Shared contract, Section 1)
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
The line form is exact: `FLAKY: <package path>.<TestName> (failed <k> of <n>)`, with the package path relative to the module, such as `internal/tui`.
The helper is a dev tool of this repo, not an `acta` subcommand, and it runs inside the existing `acta run-one` lock.
Runs without `--full` do not change. Every new test package calls `testguard.Watch()`.

## Waves

Wave 1: Task 1. Wave 2: Task 2.

### Task 1: rerun logic over go test -json

**Files:** Create `internal/flakerun/flakerun.go`, `internal/flakerun/flakerun_test.go`, `internal/flakerun/main_test.go`.
**Seam:** `flakerun.Run(args []string, run func(args []string) (io.Reader, error), out io.Writer) int`, where `run` returns a `go test -json` event stream and the result is the exit code.
**verify:** The exit code is 0 only when every failed test passed on a rerun, and every such test gets one FLAKY line; on every other path it is 1 and no FLAKY line claims a pass. Cases: one test fails then passes (FLAKY, exit 0); a test fails 3 of 3 (exit 1, named); a failed subtest reruns its top-level test; a build failure, a panic or timeout with no test name, and more than 5 failed tests rerun nothing and exit 1; reruns keep the first run's flags (`-tags`, `-short`) and add `-run '^(TestX)$' -count=1` for that package only; test output prints as plain `go test` text, not JSON.
- [ ] Red: table tests over recorded event streams fail because `flakerun.Run` does not exist.
- [ ] Change: Parse the events with `encoding/json`, collect failed top-level tests per package, rerun through `run`, print output and FLAKY lines, then run `scripts/test ./internal/flakerun/` to green.
- [ ] Commit: `feat(flake): rerun failed go tests and report flakes`.

### Task 2: scripts/test --full uses the helper

**Files:** Create `cmd/flakerun/main.go`; modify `scripts/test`, `scripts/test_test.go`, `.acta/wiki/test-locked-runs.md`.
**Seam:** `scripts/test --full` as `scripts/test_test.go` drives it.
**verify:** Wiring only: every `--full` path (whole repo, named packages, the `notui` run) goes through `go run ./cmd/flakerun` inside `acta run-one`, and every path without `--full` maps to the same `go test` call as before; the script still exits with the helper's code. The wiki page names the rerun and the FLAKY line.
- [ ] Red: `TestScriptMapsEachCallToOneGoCall` fails because its new `--full` rows expect the `flakerun` call.
- [ ] Change: Add `cmd/flakerun/main.go` that runs `go test -json` through `os/exec` and calls `flakerun.Run`, switch the three `--full` calls in `scripts/test`, update the wiki page, then run `scripts/test ./scripts/` to green.
- [ ] Commit: `feat(flake): scripts/test --full reruns flaky tests`.

## Build brief

Land order: PLN-0204 and PLN-0205 land on main in either order; they share no file. PLN-0203 lands after PLN-0204.
