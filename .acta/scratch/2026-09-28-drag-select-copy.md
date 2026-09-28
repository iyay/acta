---
id: SCRATCH-3
hash: lthv
title: drag-select-copy
status: raw
created: "2026-09-28"
---
# Drag to select text in any pane, copy on release

User, 2026-09-27: text can be selected in every pane, inside the TUI itself (mouse drag, not a terminal shift-drag bypass), and the selection copies to the clipboard automatically on release, like herdr.

Likely route: app-side drag selection kept inside one pane, copy through OSC 52.

Known gotcha: the TUI runs with tea.WithMouseCellMotion (internal/cli/cli.go), which captures the mouse, so the terminal cannot select text; most terminals allow shift/option+drag as a bypass.
