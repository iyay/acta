---
id: SPC-0066
created: "2026-10-01 18:59:35"
hash: nuwdzo5
started: "2026-10-01 19:25:29"
finished: "2026-10-01 19:57:45"
---
Status: approved by the user on 2026-10-01 (Bounded).

# Zoom the detail pane, focus a pane with the wheel, short copy toast

Three small changes to the TUI in `internal/tui`.

## 1. `z` zooms the detail pane

Today `toggleExpand` (`internal/tui/sidebar.go:322`) returns early when the detail pane has the focus, so `z` does nothing there.

- `z` on the detail pane sets `expanded` to `paneDetail`. A second `z` sets it back to `-1`.
- While the detail pane is zoomed, the layout takes the same path as a screen below 60 columns (`view.go:127`, `g.full`): the detail pane fills the whole body and the left column is not drawn. The mouse reads the same `geom`, so clicks, the wheel and drag-select land where they look.
- `focusPane` already resets `expanded`, so moving the focus (`tab`, a click) also ends the zoom.
- `leftHeights` must keep treating `expanded == paneDetail` as "nothing expanded" in the column.
- The help line `z  expand the focused pane` stays as it is.

## 2. The wheel moves the focus

This replaces the ruling in `2026-09-27-tui-polish-design.md` line 60 ("a wheel over any other pane does nothing, no focus change"). User ruling 2026-10-01: option (b), focus on wheel, not on hover.

- In `mouse()` (`internal/tui/model.go`), a wheel notch over a pane that is not focused calls `focusPane(p)` and then scrolls that pane by the notch, the same as a notch on a focused pane.
- A wheel over no pane (top bar, status line, a pane the tab does not have) stays ignored.
- A list box that was expanded goes back to normal when the wheel focuses another pane, the same as a click does today.
- Mouse mode stays `tea.WithMouseCellMotion()`. No hover tracking, no extra events.

## 3. Copy toast says "copied to clipboard"

- `copyID` (`model.go`, the `y` key) sets the status to `copied to clipboard` instead of `copied <id>`.
- `copyPicked` (`select.go`, drag-select) sets the same `copied to clipboard` instead of `copied <text>`. The truncation of the picked text goes away with it.
- `copy failed: <err>` and `nothing selected` stay as they are.

## Testing

Each change starts with a red test in `internal/tui`.

- Zoom: with the detail pane focused, `z` gives a frame where the detail box spans the full width and no left column shows; a second `z` gives the old frame back.
- Wheel: a notch over an unfocused pane moves the focus there and scrolls that pane one step. The old test that says a wheel over another pane changes nothing is flipped.
- Toast: after `y` and after a drag-select copy, the status is exactly `copied to clipboard`.
