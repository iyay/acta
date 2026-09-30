---
created: "2026-09-30"
id: SPC-0046
hash: nob0t94
started: "2026-09-30"
---
# Clickable top tabs, a lazygit-wide sidebar, and key hints for the focused pane

Status: design approved by the user in chat on 2026-09-30. Bounded: the tab bar, the mouse handler, the left column width and the status line all exist in `internal/tui`.

## Why

1. The tab bar at the top (`tabBar` in `internal/tui/view.go`) cannot be clicked. `hit` (`view.go:292`) only looks at the pane boxes, so a click on the bar does nothing.
2. The left column is `clamp(width*3/10, 28, 48)` (`internal/tui/frame.go:22`). On a wide screen it stops at 48 columns. lazygit gives its side panels a third of the width (`sidePanelWidth: 0.3333`) with no top limit.
3. The left of the status line always says `? help` (`hints`, `view.go:56`). The reader has to open the help to learn what the focused pane can do. lazygit shows the keys of the focused panel on its bottom line.

## Design

1. **Click a top tab.** `tabBar` also gives the x span of each name it draws, using the same drop logic, so the drawn names and the click spans cannot drift apart. A left press on any of the `barRows` top lines that lands on a name calls `m.openTab(i)`, the same as keys `1`-`6`. A press on the space between names does nothing. The open popup, help, new-bug and search guards at the top of `mouse` still come first.
2. **Sidebar width.** The left column becomes `max(width/3, 28)`. The 48 cap is dropped. Below 60 columns the narrow layout stays as it is.
3. **Key hints.** The fixed `hints` string becomes `m.hints()`, built from the focused pane, the open tab and the selected row. Each hint reads `Label: key`, joined by ` | `, the way lazygit does it.
   - List and Done panes: `Detail: enter`, `Status: s`, `Type: t`, `Edit: e`, `Copy id: y`, `New bug: n`, `Sort: o`. The Plans tab adds `Fold: space`. The Done pane adds `Done tab: [ ]`.
   - Detail pane: `Scroll: j k`, `Status: s`, `Edit: e`, `Copy id: y`, `Back: esc`.
   - When the selected row is a task or a debt line, `Tick: +` and `Untick: -` come first, and `Status: s` and `Type: t` are left out, because the popup refuses those rows.
   - `Help: ?` is always last. When the line is too narrow, hints drop from the right, but `Help: ?` stays.
   - Search text and status messages still take the place of the hints, as they do now.
   - The hint words keep the faint brush the old `? help` had.

## Order

Build this after PLN-0052 (the `+` and `-` keys, in `../acta-tui-tick-items`) lands. Both change `internal/tui/view.go`, and the Tick hints need those keys to exist.

## Testing

1. A click on each tab name opens that tab. A click on the name of a tab kept on a narrow screen opens it. A click between names leaves the tab as it was.
2. The left column width at 90, 150 and 240 columns is `max(width/3, 28)`. Old tests that expect 30% or 48 get updated.
3. The hints for the list pane, the Done pane, the detail pane, the Plans tab and a task row. On a narrow screen the hints are cut, and `Help: ?` is still there.
