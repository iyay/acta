---
id: SCR-0046
hash: hv9qzi7
title: 'Flaky tests: parallel tests swap the global clock, and one TUI click test'
status: raw
created: "2026-10-05 19:48:18"
schema: "1"
finished: "2026-10-05 20:11:40"
---
# Flaky tests: parallel tests swap the global clock, and one TUI click test

## Words

### 2026-10-05

Seen 2026-10-05 at the PLN-0092 land: one full run failed two tests that pass alone and in two more full runs.

- internal/write/subject_test.go TestNewScratchCommitsChoreScope calls t.Parallel() and then fixNow(t), which swaps the package global Now. Another parallel test restoring Now in its cleanup gives this test today's date: want "chore(scratch): new 2026-09-26-newest-first", got "... 2026-10-05-...". Any parallel test that calls fixNow or fixNowAt has the same race.
- internal/tui TestClickOnTheBarTakesTheHighlightOff (tabbar_test.go:92) failed once with "the screen still shows the band"; cause not checked.

Idea, not decided: tests that swap Now must not run in parallel, or Now becomes a field the test passes in.

## Context

## Log

### 2026-10-05

2026-10-05 debug of the TUI flake, no root cause yet:
- TestClickOnTheBarTakesTheHighlightOff alone: 300 runs at -parallel 8, all pass. With TestThumbSitsOnTheBorderNotInside (parallel, flips the global lipgloss color profile): 40 runs pass. With TestPaintDragOnlyTouchesTheSpans (same): 40 runs pass. Full package -race x4 in a clean clone: pass. It failed twice on main: once in a full run, once in `-race -count=2`.
- Ruled out: the status clock (View reads m.now, not time.Now); the two profile-flipping parallel tests (not reproduced, still a hazard).
- Lead not checked: the failing dump showed "copied to clipboard" and a 40-row plans board, which matches the drag tests' state; frameCache is a pointer shared by model copies (model.go:81-90), so a cache shared across tests would explain a foreign frame. longModel builds a fresh model per test, so this needs a trace of which frame View returned.

### 2026-10-05

User ruling 2026-10-05: leave the TUI flake for now; when it fails again in a land, capture the frame (shown vs want) right then.

## Open questions
