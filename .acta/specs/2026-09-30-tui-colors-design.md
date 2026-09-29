---
created: "2026-09-30"
id: SPC-0030
hash: rw80yso
---
# TUI colors

Status: design approved by the user in chat on 2026-09-30. Bounded (the theme code and every pane already exist; this changes how they paint).

## Why

The TUI uses only 3 of the 16 theme slots: slot 12 for the accent, slot 4 for work under way, slot 8 for the dim behind a popup (`newStyles`, `internal/tui/styles.go`). Every other piece of text is the theme foreground or faint. Rows the cursor is not on are faint, so the screen looks flat. The user wants the screen to use the theme's colors, the way an editor with the same theme does.

## Rule for every color

Every new color is a slot number, never a new hex. The theme gives the hex for that slot, and the `terminal` theme gives the plain ANSI number, the same as `slot` in `newStyles` today. So every built-in theme and every `~/.acta/themes/<name>.yaml` works with no change to the theme format.

## Changes

1. **Kind colors.** Each kind has one slot: scratch green (2), bug red (1), debt yellow (3), spec magenta (5), plan blue (4). The Activities tab uses cyan (6). The short id at the start of a list row, the id line in the detail, and the tab names all use these colors.
2. **Top tabs.** A tab that is not active shows its name in its kind color, not faint. The active tab is a band: the kind color as background, the text in the theme background color, bold. In the `terminal` theme, which has no background color, the band text uses slot 0.
3. **Rows are not faint.** List rows and the work lines in the detail are drawn at normal brightness: the id in the kind color and the title in the foreground color. The row under the cursor keeps the selection band and is bold. Rows with work under way keep the `work` color (slot 4), but not faint. Help text, hints and the pane borders that do not have the focus stay faint.
4. **Detail.** The labels (`ID`, `STATUS`, `FROM` and the rest) are cyan (6). The status dots: done `✓` green (2), under way `●` the accent, waiting `○` grey (8). Problem lines that start with `!` are red (1).
5. **Detail footer.** The `created · started · finished` line gets a `─` line above it, drawn the same way as the line under the header. The three labels are magenta (5), and the dates stay in the foreground color. The sticky footer grows from one line to two. The rule for when the pane is too short for a sticky header and footer counts both lines.
6. **Status line.** The word `live` is green (2). `acta` uses the accent. The clock and the version stay faint.
7. **Copy toast hides.** After `y` copies an id, the status line shows `copied <id>` as it does today. Two seconds later a `tea.Tick` message clears it. The message carries the text it is meant to clear, so when a newer message has replaced it by then, the newer one stays. `copy failed: ...` and `nothing selected` do not hide.

Unchanged: the selection band colors, the focus border, the popup dim, and the theme file format.

## Files

`internal/tui/styles.go` (the new brushes and the kind-to-slot map), `internal/tui/view.go` (tabs, rows, status line), `internal/tui/detail.go` (labels, dots, problems, footer line), `internal/tui/model.go` (the toast tick).

## Testing

Each change gets a red test first.

- `styles_test.go`: each kind and each role maps to the slot named above, for `tokyo-night` (hex) and for `terminal` (ANSI number).
- View tests: a list row that is not selected has no faint code and carries the kind color on its id; the selected row is bold; the active tab has the kind color as background; a tab that is not active has it as foreground.
- Detail tests: the labels, dots and `!` lines carry their colors; the footer is two lines with the `─` line first; the short-pane rule still falls back to one scrolling block when the pane cannot fit header, both footer lines and 3 middle lines.
- Toast tests, sending the messages straight to `Update`, with no sleep: the clear message for `copied X` empties the status; a clear message whose text no longer matches leaves the newer status alone; a failed copy does not start a tick.
