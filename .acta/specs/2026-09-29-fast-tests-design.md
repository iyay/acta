---
id: SPEC-25
hash: t539
---

# Faster Test Suite

Status: approved by the user on 2026-09-29.

## Why

`go test -count=1 ./...` takes 1:41. `internal/tui` alone takes 94s. Three tests in `internal/tui/view_test.go` take 82s of that, because they loop over every terminal size:

- `TestViewNeverOverflowsAnyWindow` (43s): width 30..200 times height 10..60.
- `TestNoRoundedCorners` (22s): width 1..200 times every tab and pane.
- `TestThumbSitsOnTheBorderNotInside` (17s): tab times pane times scroll spot.

No test in the repo calls `t.Parallel()`. Packages that run git (`cmd/acta` 28s, `internal/cli` 22s, `internal/write` 26s) run one test at a time.

## Design

1. **Sampled sizes under `-short`.** When `testing.Short()` is true, the three tests loop over a fixed list of edge sizes instead of every size: the smallest sizes (1, 29, 30, 31), the layout breakpoints read from the tui code, common sizes (80x24, 120x40) and the largest (200, 60). Without `-short` the full sweep runs as today.
2. **Daily vs gate.** Daily runs use `go test -short ./...`. The `acta:land` gate stays `go test ./...`, so the full sweep still guards every merge.
3. **Parallel tests.** Add `t.Parallel()` to top-level tests that do not call `os.Chdir`, `t.Chdir`, `t.Setenv`, `os.Setenv` or touch package-level state. Check file by file. Tests that do any of these stay serial (Go forbids `t.Setenv` in parallel tests anyway).

## Out of scope

- Changing what any test checks.
- Production code.

## Verify

- `go test -count=1 ./...` passes.
- `go test -count=1 -race -short ./...` passes, to catch shared state exposed by parallel runs.
- Report wall time before and after for both `-short` and full runs. Targets: `-short` about 30s, full about 60s.

TDD note: this change edits tests only, so there is no new red test. Proof is a green suite, the race run and the timings.
