---
parent: debt/2026-09-30-omp-eval-runner
depth: minimal
closes: [SPC-0092, DBT-0048.01, DBT-0048.02, DBT-0048.03, DBT-0048.07, DBT-0048.08, DBT-0048.09]
id: PLN-0101
created: "2026-10-06 08:11:28"
hash: wo20ff9
started: "2026-10-06 08:14:58"
---
# omp eval runner safety Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Tests in `internal/evalomp` can never touch the real HOME, `--case` fails loudly when it runs nothing, and no scaffold, omp or judge child can hang a run or outlive it.

**Spec:** .acta/specs/2026-10-06-omp-eval-runner-safety.md

**Tests:** `scripts/test ./internal/evalomp/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Tests use the fake omp in `internal/evalomp/run_test.go` (`fakeOmp`); no test runs the real omp.
- Run tests with `scripts/test`, never bare `go test`.
- `RunAll` keeps its signature `RunAll(cases []Case, o Options, only string, judge Judge, out io.Writer) (failed bool)`; `internal/cli/eval_omp.go` already exits 1 when it returns true.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02
- Wave 3: Task 03
- Wave 4: Task 04
- Wave 5: Task 05
- Wave 6: Task 06

Every task from 02 to 05 touches `internal/evalomp/run.go` and `run_test.go`, so they run one after another.

### Task 01: Tests never see the real HOME

**Files:** `internal/evalomp/main_test.go`, `internal/evalomp/run_test.go`

**verify:** No test in the package can read or write the user's real HOME, on any path, and `TestRunCaseScaffold` fails whenever `RunCase` stops giving the scaffold the throwaway home. List every test that runs a scaffold or omp and how each gets its HOME.

- [x] Red: in `TestRunCaseScaffold`, set `HOME` to a fresh `t.TempDir()` and assert `$HOME/.acta/config.yaml` does not exist after `RunCase`; watch it fail by dropping the `withEnv(..., "HOME", home)` on the scaffold command for one run, then put it back.
- [x] Green: `TestMain` sets `HOME` to a fresh temp folder (made with `os.MkdirTemp`, removed after `m.Run()`) before any test runs, so later tests are safe too.
- [x] Commit: `test(evalomp): keep every test off the real HOME (DBT-0048.01)`

### Task 02: --case fails when it runs nothing

**Files:** `internal/evalomp/run.go`, `internal/evalomp/run_test.go`

**verify:** `RunAll` never reports success when the glob is malformed or matches no case: it runs no case, prints one line naming the glob, and returns true. An empty glob still runs every case. List every glob shape tested (empty, matching, malformed, matching none).

- [x] Red: `TestRunAllBadCase` with `only` = `"[abc"` and `only` = `"nope*"`; each must return true, print a line that contains the glob, and run no case (the fake omp's `args` log file is never written).
- [x] Green: before the loop, check `only` once with `path.Match(only, "")`; on `path.ErrBadPattern` print `bad --case glob "<glob>": <err>` and return true; after a first pass that counts matches, zero matches prints `no case matches "<glob>"` and returns true.
- [x] Commit: `fix(evalomp): --case with a bad or empty match fails the run (DBT-0048.02, DBT-0048.03)`

### Task 03: The scaffold has a time limit

**Files:** `internal/evalomp/run.go`, `internal/evalomp/run_test.go`

**verify:** No scaffold can run longer than its case's `TimeoutSeconds`: past it, `RunCase` returns an error that says the scaffold timed out, and the throwaway folder is still cleaned by the caller. List every scaffold exit path (ok, fails, times out).

- [ ] Red: add a "scaffold timeout" case to `TestRunCaseFailures` with a scaffold `exec sleep 20` and `TimeoutSeconds: 1`; it must return within a few seconds with an error containing `scaffold` and `timed out`.
- [ ] Green: run the scaffold with `exec.CommandContext` under a `context.WithTimeout` of `TimeoutSeconds`, set `WaitDelay`, and on `context.DeadlineExceeded` return `scaffold: timed out after <n>s`.
- [ ] Commit: `fix(evalomp): stop a scaffold that runs past the case timeout (DBT-0048.07)`

### Task 04: A timed-out omp takes its children with it

**Files:** `internal/evalomp/run.go`, `internal/evalomp/run_test.go`

**verify:** After `RunCase` returns `ErrTimeout`, no process omp started is still alive. List every process the test starts and how it checks each is gone.

- [ ] Red: `TestRunCaseTimeoutKillsChildren` with a fake omp body `sleep 30 & echo $! > "<logdir>/child"; wait`, `TimeoutSeconds: 1`; after `ErrTimeout`, `syscall.Kill(pid, 0)` on the saved child pid must report the process is gone (allow a short poll, under 2 s).
- [ ] Green: start omp with `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` and `cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }`; keep `WaitDelay`.
- [ ] Commit: `fix(evalomp): kill omp's whole process group on timeout (DBT-0048.08)`

### Task 05: The judge reports why it failed

**Files:** `internal/evalomp/run.go`, `internal/evalomp/run_test.go`

**verify:** Every judge failure carries omp's stderr in its error, and a judge child holding the pipes open cannot keep the call waiting past `WaitDelay`. List every judge exit path (ok, omp fails, times out).

- [ ] Red: `TestOmpJudgeFailureSaysWhy` with a fake omp body `echo judge-broke >&2; exit 3`; the error must contain `judge-broke`.
- [ ] Green: in `OmpJudge` set `cmd.WaitDelay = 5 * time.Second`, capture stderr in a buffer, and on error return `fmt.Errorf("judge: %v: %s", err, strings.TrimSpace(stderr.String()))`.
- [ ] Commit: `fix(evalomp): judge errors carry omp's stderr (DBT-0048.09)`

### Task 06: Version 0.1.24

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.24, and `internal/plugincheck` passes. List each file and the version it holds.

- [ ] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [ ] Green: bump the patch from 0.1.23 to 0.1.24 in all three files.
- [ ] Commit: `chore: version 0.1.24`
