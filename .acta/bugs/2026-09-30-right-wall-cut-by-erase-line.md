---
id: BUG-0015
hash: dqy5od2
fixed_in: "1641042"
finished: "2026-09-30"
---
# The right wall of every pane is missing

## Symptom
With a theme that has its own background (the default tokyo-night), the last column of the screen is always blank. The right border of every pane and of the tab box is gone, and the last letter of the version on the status line is cut (`…97501dc307b` instead of `…97501dc307b8`). Seen on main 97501dc in herdr on Ghostty, 2026-09-30.

## Root cause
`paintFrame` (internal/tui/styles.go) adds `\x1b[K` (erase to end of line) after every line, so the theme background reaches the right edge. The view pads every line to the full window width, so after the last cell the cursor sits in the last column, waiting to wrap. In that state, erase-line clears the cell under the cursor, which is the last cell of the line. Bubble Tea's own renderer skips erase-line on full-width lines for this reason.

Proof:
- `stty -f` on the TUI's tty gives 58x204.
- `herdr pane read --format ansi` of the user's TUI pane shows 204 cells on every line, with cell 204 blank on every line.
- In a fresh herdr pane, `printf "%sA\033[K\n%sB\n"` with a run of (cols-1) dashes loses the `A` and keeps the `B`.

## Repro
1. Run `acta` with the tokyo-night theme at any window size.
2. Look at the right edge: no `┐`, `│` or `┘`, and the last character of the status line is missing.

## Found in
main 97501dc. Found by acta:debug from the user's screenshot, 2026-09-30.
