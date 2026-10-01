---
parent: scratch/2026-10-01-test-runs-share-cpu
id: SPC-0062
created: "2026-10-01 15:50:25"
hash: ojeeitr
---
Status: approved by the user on 2026-10-01 (Architectural), section by section.

# Stop parallel and orphaned test runs from pinning the CPU

Fixes `bugs/2026-10-01-parallel-test-runs-pin-cpu`.

## Problem

On 2026-10-01 every core sat at 100% (load average 281). Three gaps added up:

1. `scripts/test` takes the `acta run-one` lock only for `--full`. Short runs of the whole repo from every worktree run at the same time.
2. Agents run bare `go test` and skip `scripts/test`. The omp logs for that day hold 54 such commands out of 230, none with `-short` and none under the lock. Some were `go test ./... -count=1 -timeout 1800s`.
3. When the `go` process is killed, its `pkg.test` child keeps running with a new parent (launchd). It runs until the package ends or `-timeout` fires, up to 30 minutes.

Each orphan adds load. Later runs get slower, time out, get killed, and leave more orphans.

## Design

### 1. `scripts/test` locks whole-repo runs

- No package given, or `./...` given: run `go run ./cmd/acta run-one -- go test -short ... ./...`. This is the same lock `--full` uses, so only one whole-repo run uses the machine at a time, short or full.
- A package given (`scripts/test ./internal/write -run X`): stay `exec go test -short ...` with no lock, so the TDD loop never waits.
- `--full` does not change.
- While waiting, the line `waiting for another run-one to finish` still prints, so the agent knows it is queued and not hung.

Tests in `scripts/test_test.go` cover the three cases: a whole-repo run goes through `run-one`, a package run does not, and `--full` is the same as before.

### 2. `pre-tool` blocks bare `go test`

- Active only when the git top-level of the hook's working folder has `scripts/test`. Other repos that use the acta plugin see no change.
- Blocks `go test` when it is a real command: at the start, after `&&`, `||`, `;`, `|` or `(`, or after a prefix such as `rtk`, `env X=1` or `time`. Any package and any flags.
- Lets through: plain text such as `grep "go test" f` or `echo go test`, any `scripts/test ...`, and `acta run-one -- go test ...` (it already holds the lock).
- Blocks the same way the brainstorm block does: exit 2, reason on stderr. Claude Code (`hooks.json`) and omp (`tool_call` in `plugin/omp`) both already use this path, so `plugin/omp` does not change.
- Message, with the package and `-run` taken from the blocked command:
  `acta: run tests with scripts/test, not go test. It adds -short and the machine-wide lock. Use: scripts/test ./internal/tui -run X`
- Fails open. If the top-level cannot be found, `scripts/test` cannot be read, or the command cannot be parsed, the call goes through. A broken hook must never block every Bash call.
- `plugin/skills/build/SKILL.md` and the brief in `plugin/skills/build/dispatch.md` each get one line: never run bare `go test`; the hook blocks it. Plugincheck must stay green.

A table test in `internal/hook` covers blocked commands, plain-text matches that pass, `scripts/test` and `run-one` that pass, and a repo with no `scripts/test` that passes.

### 3. Test binaries stop when their parent dies

- New package `internal/testguard` with one function, `Watch()`. It saves `os.Getppid()` at start. A goroutine checks it every second. When it changes, it prints `testguard: parent process is gone, stopping` to stderr and calls `os.Exit(1)`. It checks for a change, not for ppid 1, so a Linux subreaper is covered too.
- Every package with tests calls it. There are 15 today. The 11 with no `TestMain` get `func TestMain(m *testing.M) { testguard.Watch(); os.Exit(m.Run()) }`. The 4 that have one (`cmd/acta`, `internal/trees`, `internal/tui`, `internal/write`) get `testguard.Watch()` as their first line.
- A test in `internal/plugincheck` walks every folder with a `*_test.go` and fails when one does not call `testguard.Watch()`. New packages are caught that way.
- `Watch`'s own test starts its binary as an orphan with `sh -c '<binary> -test.run TestChild & echo $!'`. `sh` exits at once. The test waits up to 5 seconds and checks that the child pid is gone. This is the repro from the bug file.

The cost is one goroutine per test binary, waking once a second.

## Rollout

- The hook calls the `acta` on PATH. The block works only after land and `go install ./cmd/acta`. The land report says so.
- omp runs the hook on every call, so it picks the new binary up at once. Skill text shows up only after a session restart.
- The eval sandbox has no `scripts/test`, so it is not blocked. The diff touches `plugin/skills/`, so `scripts/eval` runs at land.

## Out of scope

- Other languages (npm, cargo, pytest).
- Limiting `-p` or `GOMAXPROCS`.
- Killing orphans that are running now.
- Changing `acta run-one`.
