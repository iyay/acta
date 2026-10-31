---
parent: specs/2026-10-31-flake-tracker-design
depth: minimal
id: PLN-0205
created: "2026-10-31 17:59:57"
hash: fr788tx
started: "2026-10-31 18:02:30"
finished: "2026-10-31 18:12:29"
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
- [x] Red: table tests over recorded event streams fail because `flakerun.Run` does not exist.
- [x] Change: Parse the events with `encoding/json`, collect failed top-level tests per package, rerun through `run`, print output and FLAKY lines, then run `scripts/test ./internal/flakerun/` to green.
- [x] Commit: `feat(flake): rerun failed go tests and report flakes`.

### Task 2: scripts/test --full uses the helper

**Files:** Create `cmd/flakerun/main.go`; modify `scripts/test`, `scripts/test_test.go`, `.acta/wiki/test-locked-runs.md`.
**Seam:** `scripts/test --full` as `scripts/test_test.go` drives it.
**verify:** Wiring only: every `--full` path (whole repo, named packages, the `notui` run) goes through `go run ./cmd/flakerun` inside `acta run-one`, and every path without `--full` maps to the same `go test` call as before; the script still exits with the helper's code. The wiki page names the rerun and the FLAKY line.
- [x] Red: `TestScriptMapsEachCallToOneGoCall` fails because its new `--full` rows expect the `flakerun` call.
- [x] Change: Add `cmd/flakerun/main.go` that runs `go test -json` through `os/exec` and calls `flakerun.Run`, switch the three `--full` calls in `scripts/test`, update the wiki page, then run `scripts/test ./scripts/` to green.
- [x] Commit: `feat(flake): scripts/test --full reruns flaky tests`.

## Build brief

Land order: PLN-0204 and PLN-0205 land on main in either order; they share no file. PLN-0203 lands after PLN-0204.

## Fix round 1

### Task 3: A crashed test binary is never a flake

**verify:** A package whose test binary died (a test with a run event but no pass, fail or skip event, or panic output in the package) reruns nothing and the run exits 1, even when another test in that package failed by name; a rerun counts as a pass only when go exits 0 for it.

- [x] internal/flakerun/flakerun.go failedWithNoTest: one failed test hides a later timeout or crash in the same package, so the earlier test reruns, passes, prints FLAKY and the run exits 0 though the rest of the package never ran (TestA fails once, TestB hangs past -timeout: FLAKY TestA, exit 0). Treat a package with a started test that has no end event, or panic output, as crashed: rerun nothing there, exit 1.
- [x] A named panic kills the binary, so later tests in that package never run; rerunning only the panicked test must not turn the run green. Same crash rule.
- [x] internal/flakerun/flakerun_test.go: add the regression case, a failed TestA plus a timeout in the same package, expecting exit 1 and no FLAKY line; and a named panic case.
- [x] internal/flakerun/flakerun_test.go: add a rerun row where the test shows pass but run returns an error (TestMain exits non-zero), expecting no FLAKY line and exit 1, so the err == nil check in rerun is pinned.
- [x] Commit with the trailer `Task: PLN-fr788tx#3`

## Review notes

- Review: deep level, no floor raised; round 1 BLOCKED (a crash hidden by an earlier named fail, fact-check kept it), fix round 1, round 2 CLEAN, no polish. Notes kept, not tracked: CombinedOutput buffers output until go test ends; go run adds an exit status line on red; flagsOf drops a ./ flag value; the module-root package keeps its full path; no test pins the script exit code; any panic is red, never a flake; the crash message can name a later package than the first that died; a goroutine os.Exit between two tests is a tiny unseen window.

## State

### Next

fix round 1 done (a5b30e9: crashed binary reruns nothing); round 2 review next
