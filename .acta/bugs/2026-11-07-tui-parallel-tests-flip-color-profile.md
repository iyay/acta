---
id: BUG-0079
hash: c2c89mu
---
# Parallel TUI tests flip the global color profile and break each other

## Symptom
`internal/tui.TestClickOnTheBarTakesTheHighlightOff` (BUG-0058) fails now and then in the full suite with "the screen still shows the band", and its frame holds `[2m` color codes. `TestReviewRowWearsItsOwnColor` fails 149 of 150 runs when run with the other parallel color tests, and 0 of 150 alone.

## Root cause
`withColors` (`internal/tui/view_test.go:27-28`) and `withTrueColor` (`internal/tui/styles_test.go:21-22`) call `lipgloss.SetColorProfile`, which is process-wide. Three tests that use them are `t.Parallel()`: `TestPaintDragOnlyTouchesTheSpans`, `TestThumbSitsOnTheBorderNotInside`, `TestReviewRowWearsItsOwnColor`. Any parallel test that renders while one of them holds a color profile, or resets it to Ascii, sees the other profile. `TestClickOnTheBarTakesTheHighlightOff` compares two `View()` calls (`internal/tui/tabbar_test.go:89-92`), so a flip between them makes them differ.

## Repro
`scripts/test ./internal/tui/ -run '^(TestClickOnTheBarTakesTheHighlightOff|TestPaintDragOnlyTouchesTheSpans|TestReviewRowWearsItsOwnColor)$' -count=150` fails `TestReviewRowWearsItsOwnColor` 149 times. A probe that sets `termenv.ANSI256` between the two `View()` calls of the BUG-0058 test fails with the same assertion; the same probe without the flip passes.

## Found in
main, debug of BUG-0058 on 2026-11-07. The natural interleaving of the BUG-0058 test itself was not caught in 320 paired runs.
