---
parent: scratch/2026-09-30-faster-agent-test-runs
created: "2026-09-30"
id: SPC-0033
hash: zssi5lc
status: approved
---
# Faster Test Runs for Agents

Status: design approved by the user in chat on 2026-09-30, section by section. Architectural (a new CLI command, and a rule change in the build, plan, dispatch, land and review skills that every repo using acta follows).

## Why

- SPC-0025 landed, but its targets were missed. Measured 2026-09-30 on 8 cores and 16 GB: `go test -count=1 -short ./...` takes 67s, and the full `go test -count=1 ./...` takes 100s.
- One plan runs the full suite many times. The build baseline (`plugin/skills/build/SKILL.md:116`) runs it once, each task verify line runs it again (62 plan lines say `go test ./...`), the end of the build (`build/SKILL.md:208`) runs it, the land precondition runs it, and so does the run after the merge. A 5-task plan comes to about 9 full runs.
- Agents run at the same time. At 06:15 three worktrees ran `go test ./...` at once. The load average was 26, only about 128 MB of RAM was free, and the machine was swapping. One suite stretched to 3-5 minutes.
- The full suite is slow even alone. `internal/tui` takes 57s of wall time but keeps only about 2 of 8 cores busy, because three sweep tests are single serial loops (41s, 24s, 14s). `t.Parallel()` runs whole tests side by side, so the package can never finish before its slowest test. The git packages barely run in parallel (cmd 0/28, cli 2/104, write 20/120), because their helpers call `t.Setenv`, and Go refuses `t.Parallel()` in a test that does that.

## Design

### 1. One full run at a time: `acta run-one`

- `acta run-one -- <cmd> [args...]` takes an OS lock (`flock`) on `<user cache>/acta/locks/run-one.lock`, runs the command, and lets go when it ends.
- There is one lock per user per machine, not one per repo. What the runs fight over is the machine's CPU and RAM.
- It reuses `lockDir` and `safeDir` from `internal/write/tick.go`, so the same folder checks apply: the folder must be absolute, owned by the user, and closed to other users.
- While another run holds the lock, it waits with no timeout, and prints one line to stderr: `waiting for another run-one to finish`.
- stdin, stdout and stderr pass straight through. The exit code is the command's own exit code. If `--` or the command is missing, it exits 2 with a usage line.
- The kernel drops the lock when the process ends, even in a crash, so no lock is ever left behind.

### 2. `scripts/test` in this repo

- `scripts/test` runs `go test -short ./...`.
- `scripts/test --full` runs `go run ./cmd/acta run-one -- go test ./...`. It uses `go run` so it always matches the code in the checkout and never needs an old `acta` on PATH.
- Any other args pass on to `go test`, for example `scripts/test ./internal/tui/` or `scripts/test --full -count=1`.

### 3. Skill rules (every repo)

The skills stay stack-neutral. Go appears only in examples.

1. During a build, run the narrowest tests that prove the task: the package or file it touched. If you need something wider, use the project's fast mode. The full suite runs only in `acta:land`. When the repo has `scripts/test`, use it: no flag means the fast mode, and `--full` means the full suite.
2. `plan/SKILL.md`: each task's run steps and verify line name a narrow command (for example `go test ./internal/tui/ -run TestX`), never `./...`. The plan header names the repo's fast command and its full command.
3. `build/SKILL.md`: the baseline (line 116) and the end-of-build run (line 208) use the fast mode. `dispatch/SKILL.md:349` GATES does the same.
4. `land/SKILL.md`: the precondition runs the full suite as `acta run-one -- <full command>`. In step 6, after the merge, the full suite runs again only when `git rev-parse HEAD^{tree}` differs from the branch's tree. When the trees match, the landing report says "tree same as branch, gates reused".
5. `review/SKILL.md:19`: reviewers run only the tests that prove a finding. They never run the full suite.

### 4. A faster full suite (tests only; what they check stays the same)

1. **Sweeps split by width.** In `TestViewNeverOverflowsAnyWindow`, `TestNoRoundedCorners`, `TestThumbSitsOnTheBorderNotInside`, `TestViewFitsEveryTerminalSize` and `TestTheTabBoxHoldsItsWidthOnEveryScreen`, the outer loop becomes `t.Run(fmt.Sprintf("w%d", w), ...)` with `t.Parallel()`. Each size still builds its own `newModel(t)`, so no state is shared. Failure messages still name the size.
2. **Git identity without `t.Setenv`.** `repoWith` (`internal/write/ops_test.go:16`), `fixtureRepo` (`cmd/acta/main_test.go:40`) and the same helpers in `internal/cli` and `internal/trees` set `user.name` and `user.email` with a repo-local `git config` right after `git init`. The commits that acta's own code makes in that repo pick the config up as well. `PM_ROOT=""` is set once in `TestMain`.
3. **More `t.Parallel()`.** Add it to tests in write, cmd, cli and trees that no longer call `t.Setenv`, `os.Setenv`, `t.Chdir` or `os.Chdir`. Tests that still need `HOME`, `XDG_CACHE_HOME`, `PM_VOICE_FILE` or a working-folder change stay serial.
4. **Known race.** `internal/tui/watch_test.go:60` reads `loads` without holding `mu` (a debt from SPC-0025). More parallel tests make a race report there more likely. If `-race` trips on it, fix it in this plan.

## Out of scope

- Rewriting verify lines in plans that are already running (tui-colors, batch-author-lookup, config-file-name).
- Taking checks out of the land gate, or making the sweeps sample fewer sizes in a full run.
- Production code, except the new `run-one` command and the move of the shared lock helpers.

## Verify

- `acta run-one`: tests show that two runs never overlap and the second one waits and then runs; that the command's exit code and output come back unchanged; that a missing `--` or a missing command exits 2; and that the lock folder checks refuse a folder that is not the user's own or is open to others.
- `go test -count=1 ./...` passes, and `go test -count=1 -race -short ./...` passes.
- Report the wall time before and after, for both `-short` and full, measured on a quiet machine under `acta run-one`. Target: the full suite alone in 40s or less. If the target is missed, report the real numbers and do not loosen any check.
- The eval gate runs at land, because the diff touches `plugin/skills/`.
