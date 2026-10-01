---
id: BUG-0024
hash: cc1tx75
priority: high
status: fixed
started: "2026-10-01 15:58:01"
finished: "2026-10-01 16:15:39"
---
# Test runs from many agents pin every CPU core at 100%

## Symptom
On 2026-10-01 at 15:21 every core sat at 100% and the load average hit 281.
btop showed four `tui.test` and one `acta.test` with no `go` parent, next to a
live `go test -short ./...`, while agents in four herdr tabs worked at once.

## Root cause
Three gaps add up.

1. `scripts/test:20` runs the plain (short) path as `exec go test -short`,
   outside the `acta run-one` lock. Only `--full` (`scripts/test:18`) waits.
   So short runs from every worktree run at the same time.
2. Agents run bare `go test` and skip `scripts/test`. The omp session logs for
   2026-10-01 hold 54 bare `go test` commands out of 230 test commands, none
   with `-short` and none under the lock, for example
   `go test ./... -count=1 -timeout 1800s`,
   `go clean -testcache && go test ./... -timeout 900s` and
   `go test ./internal/tui/ -count=1`. They came from the review-note-triage
   session and the four-old-bugs `Task4ReloadScroll` subagent, both started at
   15:12. `plugin/skills/build/SKILL.md:115` and
   `plugin/skills/plan/SKILL.md:88` name the commands to use, but nothing
   stops a bare `go test`.
3. When the `go` process is killed (tool timeout or interrupt), its `pkg.test`
   child keeps running with ppid 1 until the package finishes or its
   `-timeout` fires. With `-timeout 1800s` that can be 30 minutes. Each
   orphan adds load, so later runs get slower, time out and leave more
   orphans.

## Repro
Orphan part, in an empty module with one test that sleeps 20 seconds:

    go test -count=1 . &
    sleep 6
    kill -9 <pid of go test>
    ps -Ao pid,ppid,comm | grep '\.test'   # the test binary is still alive, ppid 1

Load part: run `go test ./... -count=1` in two worktrees at once while a third
runs `scripts/test`.

## Found in
main, by debug reading of a btop screenshot, `ps`, the omp session logs under
`~/.omp/agent/sessions`, and the repro above.
