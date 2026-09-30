---
id: SCR-0003
hash: lthvpyi
title: drag-select-copy
status: brainstorming
created: "2026-09-28"
started: "2026-09-30"
finished: "2026-09-30"
---
# Drag to select text in any pane, copy on release

User, 2026-09-27: text can be selected in every pane, inside the TUI itself (mouse drag, not a terminal shift-drag bypass), and the selection copies to the clipboard automatically on release, like herdr.

Likely route: app-side drag selection kept inside one pane, copy through OSC 52.

Known gotcha: the TUI runs with tea.WithMouseCellMotion (internal/cli/cli.go), which captures the mouse, so the terminal cannot select text; most terminals allow shift/option+drag as a bypass.

User, 2026-09-30: fold in one more change. The Scratches tab and the scratch kind are green now; pick another color. Green already means done (PLN-0047).

Color option A (slot 7 grey) rejected: too close to the title FG (#c0caf5 vs #a9b1d6), the id would blend into the title.

Color option B (orange) rejected: tokyo-night yellow #e0af68 already reads as orange for Debts. Warm hues are all taken; left: teal (green+cyan) or pink (red+magenta).

Scratch color picked: #ff79c6 (hot pink, candidate 4 of the swatch). Green stays for done only.

Selection shape: stream, like a terminal (start cell to end cell, middle rows whole). Kept inside one pane; borders and scrollbar never copied.

Drag past the pane edge: no auto-scroll. Only visible text is selectable; copied text is what the screen shows (cut titles stay cut).

Approach picked: 1, screen grid. Selection = pane + start and end cell in screen coords, clamped to the box inner rect. On release, text comes from the last drawn frame (ansi.Cut + ansi.Strip, trailing spaces trimmed) and goes to copyText. Highlight is spliced onto the frame the same way. Rejected: per-pane selection (5-6 copies of the same logic), terminal bypass only.

Design section 1 approved (mouse): press records an anchor only in the text area and keeps todays click behaviour; drag moves the end cell clamped to that pane; release copies and shows the y toast; click with no drag copies nothing; highlight clears on next press, any key, wheel or resize; off while popup, help, search or slug input is open.

Toast ruling: same toast as y (copyID): m.status plus clearStatusAfter(toastFor); error "copy failed: ..." stays. Text is "copied <copied text>", newlines as spaces, cut with an ellipsis to fit the status line.

Design section 2 approved (text, highlight, scratch color, tests). Next: write the spec.
