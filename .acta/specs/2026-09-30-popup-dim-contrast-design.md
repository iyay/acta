---
parent: bugs/2026-09-30-popup-dim-turns-bright-in-ghostty
created: "2026-09-30"
id: SPC-0038
hash: gxr7lam
---
# Popup dim that survives minimum contrast, accent keys in help, and the right wall

Status: design approved by the user in chat on 2026-09-30. Bounded (the popup, `cover` and the help box all exist; this changes two brushes and one line ending). The user asked to fix BUG-0015 in the same spec on 2026-09-30.

## Why

BUG-0014: the screen behind a popup shows bright, not dim. `dim` (`internal/tui/styles.go`) is faint plus slot 8 mixed halfway toward the background. For tokyo-night that is #2d3147 on #1a1b26, a contrast ratio of 1.34. Ghostty's `minimum-contrast` (the user runs 1.5) swaps any foreground below its floor for white or black, so the dim text turns white. Faint lowers contrast further, which makes it worse.

BUG-0015: the last column of the screen is always blank, so the right wall of every pane and the last letter of the status line are gone. `paintFrame` (`internal/tui/styles.go`) adds `\x1b[K` (erase to end of line) after every line. Every line is already the full window width, so the cursor waits in the last column, and erase-line there clears that last cell.

The user also wants the key column of the `?` help popup in the theme accent, so the keys stand out from what they do.

## Changes

1. **Dim is slot 8, plain.** `dim` becomes `Foreground(slot(8))` with no faint and no mixing toward the background. In the `terminal` theme it stays ANSI `8`. Slot 8 against the theme background measures 1.85 to 4.37 across the built-in themes, all above 1.5, and it still reads as dim next to the foreground. `dimColor` and `mixHex` lose their only caller and go. `hexRGB` goes too if nothing else uses it. Their tests go with them.
2. **Help keys in the accent.** Each line of `helpLines` is split at its first run of two or more spaces. The part before (for example `1-6 ← →`, `tab shift+tab`) is drawn in `styles.accent`, and the rest stays in the foreground. The popup keeps its size and layout; only the colors change. The other popups (the value picker and the new bug prompt) do not change.
3. **Erase only short lines.** `paintFrame` takes the window width. It adds `\x1b[K` only to a line narrower than that width, so the theme background still reaches the right edge of a short line, and a full line keeps its last cell. Both calls in `draw` pass `m.width`.

Unchanged: `cover` still strips every line behind the box and paints it with `dim`. The popup box, its border and its place stay the same.

## Files

`internal/tui/styles.go` (`dim`, removal of `dimColor` and `mixHex`), `internal/tui/view.go` (the `?` popup content, the `paintFrame` calls), `paintFrame` in `internal/tui/styles.go`, with tests in `internal/tui/styles_test.go` and `internal/tui/view_test.go`.

## Testing

- For every built-in theme (every name `theme.Names()` gives, except `terminal`), the WCAG contrast ratio of the `dim` foreground against the theme background is at least 1.6. The test computes the ratio itself from the hex. This test must fail on today's code for tokyo-night (1.34).
- `dim` is never faint, in every theme.
- The `terminal` theme's `dim` is ANSI `8`, with no 24-bit code.
- `TestPopupDimsTheBackground` keeps checking every popup at both sizes and both profiles. Its brush changes to slot 8 with no faint.
- In the `?` popup, each line's key part is wrapped in the accent code and its description is not. Every line of `helpLines` is checked, and the box width is the same as before.
- On every tab, at several window sizes, with a popup open and without one: a drawn line as wide as the window never ends in `\x1b[K`, and its last cell is the character the view drew (`┐`, `│`, `┘`, or the last character of the status line). A line narrower than the window still ends in `\x1b[K`.
