---
parent: scratch/2026-10-05-parallel-test-clock-race
id: SPC-0083
created: "2026-10-05 20:11:40"
hash: xcoefvd
---
# write tests that swap the clock do not run in parallel

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: `TestNewScratchCommitsChoreScope` and `TestSubjectDropsScopeForMixedKinds` in `internal/write/subject_test.go` call `t.Parallel()` and then `fixNow`, which swaps the package global `Now`. The race detector reports a data race on `Now` (`internal/write/ops_test.go:208-210`), and parallel tests read the wrong date: one full run failed with "chore(scratch): new 2026-10-05-..." where the test wants 2026-09-26.

Design:
- Drop `t.Parallel()` from the two tests; they then run before the parallel group and the swap cannot overlap another test.
- No other test in the package calls both `t.Parallel()` and `fixNow` or `fixNowAt`.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: the TUI test `TestClickOnTheBarTakesTheHighlightOff`, whose cause is not found yet.

Tests: `scripts/test ./internal/write -race -count=3` shows no data race and no failure.
