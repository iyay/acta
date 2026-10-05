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

## Open questions
