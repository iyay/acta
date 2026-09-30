---
id: BUG-0014
hash: qu87ixv
---
# Text behind a popup shows bright instead of dim

## Symptom
With the tokyo-night theme in Ghostty, opening the help popup (`?`) or any other popup does not dim the screen behind it. The text behind the popup shows at full brightness, close to white, and every color on it is gone. Seen on main 97501dc (screenshot, 2026-09-30).

## Root cause
The TUI side is correct. `cover` (internal/tui/view.go:670) strips every line behind the popup and paints it with `styles.dim`. `dim` (internal/tui/styles.go:85) is faint plus `dimColor`, which mixes slot 8 halfway toward the theme background (internal/tui/styles.go:125). For tokyo-night that gives #2d3147 on #1a1b26, a contrast ratio of only 1.34.

The user's Ghostty config sets `minimum-contrast = 1.5`. Any cell whose foreground has less contrast than that against its background gets its foreground swapped for white or black. So the dim color becomes white, and the "dimmed" screen is brighter than the normal one.

Contrast of the mixed dim color against the background, per built-in theme: tokyo-night 1.34, tokyo-night-day 1.34, catppuccin-mocha 1.52, dracula 1.74, catppuccin-latte 1.91, gruvbox-dark 2.02. Slot 8 on its own is 1.85 or higher in every one.

Proof the bytes are right up to the terminal:
- A pty run of the real binary sends `2;38;2;44;48;71` on 48 segments after `?`.
- `herdr pane read --format ansi` of the user's own TUI pane, with the popup open, holds `2` + `38;2;44;48;71` on every cell behind the box.

## Repro
1. In Ghostty, set `minimum-contrast = 1.5`.
2. Run `acta` with the default tokyo-night theme.
3. Press `?`. The text behind the Keys box is bright.

The check: with `minimum-contrast = 1`, the same screen is dim.

## Found in
main 97501dc. Found by acta:debug from the user's screenshot, 2026-09-30.
