---
parent: bugs/2026-10-01-parallel-test-runs-pin-cpu
depth: minimal
id: PLN-0071
created: "2026-10-01 15:53:05"
hash: cgompru
---
# Test Runs Share The CPU Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Stop parallel and orphaned test runs from pinning the CPU (fixes `bugs/2026-10-01-parallel-test-runs-pin-cpu`).

**Spec:** `.acta/specs/2026-10-01-test-runs-share-cpu-design.md`

**Tests:** fast `scripts/test ./scripts ./internal/hook ./internal/cli ./internal/plugincheck ./internal/testguard`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Every run step uses `scripts/test <package> -run <Name>`, never bare `go test` and never `./...`. Task 2 makes the hook block bare `go test`.
- The hook fails open: any error, a missing git top-level, or an unreadable `scripts/test` lets the call through.
- Block message, word for word except the `Use:` part: `acta: run tests with scripts/test, not go test. It adds -short and the machine-wide lock. Use: <command>`.
- testguard stop line, word for word: `testguard: parent process is gone, stopping`.
- PLN-0069 (worktree `../acta-four-old-bugs`) also edits `scripts/test` and `scripts/test_test.go`. When it lands first, Task 1 resolves the conflict by intent: keep both changes.
- Comments are plain English a 10-year-old can read. They say why, not what.

## Waves

- Wave 1: Task 1, Task 2, Task 3, Task 4 (no shared files).
- Wave 2: Task 5 (needs `testguard.Watch` from Task 4).

### Task 1: scripts/test locks whole-repo runs

**Files:**
- Modify: `scripts/test`
- Test: `scripts/test_test.go`

**verify:** A run of the whole repo always goes through `go run ./cmd/acta run-one`, and a run that names a package never waits on the lock. List every argument shape checked (no args, flags only, explicit `./...`, `./...` with flags, one package, several packages, `--full` with and without a package) and which path each takes.

- [ ] Failing test: change the rows of `TestScriptMapsEachCallToOneGoCall` so no args, `-count=1` and `./...` expect `run ./cmd/acta run-one -- go test -short ... ./...`, and add rows for `./...` and two packages; it fails because the short path still runs `go test -short` with no lock.
- [ ] Code: in `scripts/test`, send the short path through `go run ./cmd/acta run-one -- go test -short` when no package was given or an argument is exactly `./...`; package runs keep `exec go test -short`; `--full` stays as it is; update the header comment.
- [ ] Commit: `scripts/test puts whole-repo short runs under the run-one lock`

### Task 2: pre-tool blocks bare go test

**Files:**
- Create: `internal/hook/gotest.go`
- Create: `internal/hook/gotest_test.go`
- Modify: `internal/cli/hook.go`
- Modify: `plugin/hooks/pre-tool` (comment only)
- Test: `internal/cli/hook_test.go`

**verify:** In a repo whose git top-level has `scripts/test`, no shell command that really runs `go test` gets through, and nothing else is ever blocked. List every command shape checked: at the start, after `&&`, `||`, `;`, `|`, `(`, after `cd x &&`, after `rtk`, `rtk proxy`, `env X=1`, `X=1`, `time`; plain text that passes (`grep "go test" f`, `echo go test`, `git log --grep 'go test'`); `scripts/test ...`, `acta run-one -- go test ...` and `go run ./cmd/acta run-one -- go test ...` that pass; a repo with no `scripts/test` that passes; a folder with no `.git` above it that passes.

- [ ] Failing test: a table test in `internal/hook/gotest_test.go` for `GoTestBlock(dir, command string) (bool, string)` over every shape in the verify line, checking the block flag and that the message ends in `Use: scripts/test` plus the blocked command's package args and its `-run` value; plus one test in `internal/cli/hook_test.go` that runs `acta hook pre-tool` in a temp repo with `scripts/test` and a `go test ./...` payload and wants exit 2 with the message on stderr; both fail because `GoTestBlock` does not exist.
- [ ] Code: `GoTestBlock` walks up from `dir` to the first folder holding `.git` (a folder or a worktree file), stops when that folder has no `scripts/test`, cuts the command at `&&`, `||`, `;`, `|` and `(`, strips the leading prefixes named in the verify line, and blocks a piece that then starts with `go test` unless the piece holds `run-one --` before it; the `pre-tool` case in `internal/cli/hook.go` calls it with `os.Getwd()` after the brainstorm check, prints the message to stderr and returns `exitBlock`; the top comment of `plugin/hooks/pre-tool` now says it also stops bare `go test`.
- [ ] Commit: `pre-tool blocks bare go test in repos that have scripts/test`

### Task 3: skills say never run bare go test

**Files:**
- Modify: `plugin/skills/build/SKILL.md`
- Modify: `plugin/skills/build/dispatch.md`
- Test: `internal/plugincheck/skill_build_test.go`
- Test: `internal/plugincheck/build_dispatch_test.go`

**verify:** Both the build skill and the dispatch brief tell the agent to run tests through `scripts/test` and never bare `go test`, and say the hook blocks it; no other skill text changes and every plugincheck size cap still holds. List both places checked and the cap of each file.

- [ ] Failing test: one check in each plugincheck file wants the line "never run bare `go test`" and the word "blocks" in its file; both fail because neither file says it.
- [ ] Code: add one sentence to the fast-tests paragraph of `plugin/skills/build/SKILL.md` and one to the GATES part of the brief in `plugin/skills/build/dispatch.md`: in a repo with `scripts/test`, never run bare `go test`; the pre-tool hook blocks it.
- [ ] Commit: `build skill and dispatch brief say never run bare go test`

### Task 4: testguard stops an orphaned test binary

**Files:**
- Create: `internal/testguard/testguard.go`
- Create: `internal/testguard/testguard_test.go`

**verify:** A test binary whose parent process goes away always stops within 5 seconds, and one whose parent stays alive is never stopped. List both cases and how the test proves each.

- [ ] Failing test: `testguard_test.go` has a `TestMain` that calls `Watch()`, a `TestChild` that sleeps 30 seconds only when `TESTGUARD_CHILD=1`, and a test that starts its own binary through `sh -c '<binary> -test.run TestChild & echo $!'` with that env set, reads the pid, and wants it gone within 5 seconds; a second check in the same test wants the parent test itself still running; it fails because `Watch` does not exist.
- [ ] Code: `Watch()` saves `os.Getppid()`, starts a goroutine with a one-second `time.Ticker`, and when the ppid changes prints the testguard stop line to stderr and calls `os.Exit(1)`; the comment says it checks for a change, not for ppid 1, so a Linux subreaper is covered.
- [ ] Commit: `testguard stops a test binary whose go parent died`

### Task 5: every test package calls testguard.Watch

**Files:**
- Create: `main_test.go` in `internal/board`, `internal/cli`, `internal/config`, `internal/doctor`, `internal/editor`, `internal/evalomp`, `internal/gitc`, `internal/hook`, `internal/plugincheck`, `internal/theme`, `scripts`
- Modify: `cmd/acta/main_test.go`, `internal/trees/trees_test.go`, `internal/tui/model_test.go`, `internal/write/main_test.go`
- Create: `internal/plugincheck/testguard_test.go`

**verify:** No Go test package in the repo can run without `testguard.Watch()`, including packages added later. List every folder with a `*_test.go` the check found and whether each calls it.

- [ ] Failing test: `internal/plugincheck/testguard_test.go` walks the repo from `../..`, skipping `.git`, `testdata`, `node_modules` and `plugin`, and for each folder with a `*_test.go` wants one of those files to call `testguard.Watch()`; it fails on the 11 folders with no `TestMain`.
- [ ] Code: each new `main_test.go` uses the same package clause as the other test files in its folder and holds `func TestMain(m *testing.M) { testguard.Watch(); os.Exit(m.Run()) }`; each of the 4 existing `TestMain` funcs gets `testguard.Watch()` as its first line.
- [ ] Commit: `every test package calls testguard.Watch so killed runs leave no orphans`
