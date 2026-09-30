---
created: "2026-09-30"
parent: scratch/2026-09-28-drag-select-copy
id: SPC-0042
hash: h293afb
started: "2026-09-30"
finished: "2026-09-30"
---
# TUI drag to select and copy, and a pink scratch color

Status: design approved by the user in chat on 2026-09-30, in two sections. Architectural: the TUI gets a new selection state that spans mouse input, drawing and the clipboard. The log of SCR-0003 holds every ruling.

## Why

1. The TUI runs with `tea.WithMouseCellMotion` (`internal/cli/cli.go`), so the terminal cannot select text in it. The user wants to drag inside any pane and have the text copied on release, the way herdr does it.
2. The Scratches tab and the scratch kind are green. Since PLN-0047, green also means done, so the two mix up. The user wants scratch in another color.

## Mouse

1. **Press** with the left button inside a pane keeps what a click does today: it takes the focus, moves the cursor to the row, or switches the tab. It also records that cell as the anchor. The anchor is only recorded when the cell is in the text area of the box: not the border, the title line, the bottom line or the scrollbar wall.
2. **Drag** (motion with the left button held) moves the end cell with the pointer. The end cell is clamped to the text area of the pane that holds the anchor. The selection shows once the end cell differs from the anchor.
3. **Release** with a selection copies its text at once (see Copy).
4. **A click with no drag** is a plain click. Nothing is copied.
5. **The highlight clears** on the next press, on any key, on a wheel turn and on a resize. The selection lives in screen cells, so once the content moves it would point at the wrong text.
6. **Off** while a popup, the help, search or the slug input is open, the same as the mouse is today.
7. **No auto-scroll.** A drag past the top or bottom of a pane stops at its edge. Only text on screen can be selected.

## Selection shape

A stream, like a terminal. The start is the earlier of anchor and end cell in row order, so a drag up or to the left works the same. The first row runs from the start cell to the right edge of the text area. Middle rows are whole. The last row runs from the left edge of the text area to the end cell, both ends included. A selection on one row runs from start to end.

## Copy

1. On release, the screen is drawn again with no highlight (`draw()`), and the selected rows are taken from it.
2. Each row piece is `ansi.Cut` by cell column, so wide characters stay whole, then `ansi.Strip` to drop the colors. Trailing spaces are trimmed. Rows are joined with `\n`.
3. The copied text is what the screen shows. A title cut with `…` is copied cut.
4. The text goes to the existing `copyText` (pbcopy, then OSC 52) through `m.clip`.
5. **Toast.** The same toast as `y` (`copyID`): `m.status` with `clearStatusAfter(toastFor, ...)`. A good copy says `copied <text>`, with newlines shown as spaces and the text cut with `…` to fit the status line. A failed copy says `copy failed: <error>` and stays until the next message. A selection whose text is empty after trimming copies nothing and shows no toast.

## Highlight

At the end of `draw()`, each selected row is split in three with `ansi.Cut`: the left part as it is, the middle as plain text (`ansi.Strip`) painted with the theme selection colors (`SelectionBG` and `SelectionFG`, the same fallback the `selected` brush uses), and the right part as it is. A change to the selection draws the frame again: it never reuses the frame cache.

`github.com/charmbracelet/x/ansi` is already in `go.mod` as an indirect module. It becomes a direct one. No new module is added.

## Scratch color

The Scratches tab and the scratch kind color move from green (slot 2) to pink. Pink is not one of the 16 slots, so scratch takes a fixed color instead of a slot:

- hex theme, dark: `#ff79c6`
- hex theme, light: `#c2185b`, so it stays readable on a light background
- `terminal` theme: 256-color `212`

Dark or light comes from `Theme.Dark`, the same test the selection band uses. Green stays for done, live and the pulse. Every other kind keeps its slot.

## Unchanged

A click, the wheel, the `y` key, the selection band of the cursor row, the frame reuse rule for messages that change nothing, and the theme file format.

## Testing

1. A drag in a sidebar pane and in the detail pane copies the text shown, with no border, title or scrollbar cells.
2. A drag up and to the left copies the same text as the same drag down and to the right.
3. A drag over a wide character never cuts it in half.
4. A drag past the pane edge stops at the text area of the anchor pane.
5. A click with no drag copies nothing and keeps today's click behaviour.
6. A press on the border or title line records no anchor.
7. The highlight clears on a key, a press, a wheel turn and a resize.
8. With a popup, the help, search or the slug input open, a drag selects nothing.
9. A good copy shows `copied <text>` and hides after `toastFor`; a long or multi-line text is shown on one line, cut with `…`. A failed copy shows `copy failed: ...` and stays. An all-space selection copies nothing.
10. The scratch kind color is `#ff79c6` in a dark hex theme, `#c2185b` in a light one and `212` in the `terminal` theme. Done stays green.

## Files

- `internal/tui/model.go`: selection state, the mouse handler, clearing on key, wheel and resize.
- `internal/tui/view.go`: the text area of a box, the highlight splice in `draw()`, the text taken for a copy.
- `internal/tui/styles.go`: the scratch kind color.
- `go.mod`: `github.com/charmbracelet/x/ansi` becomes direct.
- tests next to each of these.
