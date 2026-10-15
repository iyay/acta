---
id: SPC-0136
created: "2026-10-15 12:33:37"
hash: nv1w4c6
started: "2026-10-15 12:34:49"
---
# Wheel scrolls the pane under the pointer without taking the focus

Status: Bounded, approved by the user in chat on 2026-10-15.

## Problem

On the commit screen the wheel scrolls the box under the pointer and leaves everything else alone (`commitMouse` in `internal/tui/commits.go`). On the board it does not: a notch over a pane that is not focused first calls `focusPane(p)` (`internal/tui/model.go`, the wheel case in the mouse handler). That moves the border, drops the zoom (`expanded = -1`) and changes `m.last`. To read the detail pane the user has to give it the focus first, or lose their place in the list.

## Design

- On the board, a notch scrolls the pane under the pointer: the detail pane and every list pane alike.
- The focus never changes on a notch. `m.focus`, `m.last` and `m.expanded` stay as they were.
- Remove the `focusPane` branch from the wheel case. The rest stays: the `onPane` check, gathering notches per frame (`wheelDelta`, `wheelPane`, `wheelMark`) and flushing them when the pointer moves to another pane.
- A pane the tab does not draw is still ignored, as today.
- A click still moves the focus. Keys are not touched.

## Testing

In `internal/tui`, red first:
- Focus on a list pane, notch over the detail pane: the detail scroll moves and `m.focus` is still the list pane.
- Focus on the detail pane, notch over a list pane: that list scrolls and `m.focus` is still the detail pane.
- A zoomed pane stays zoomed after a notch over another pane.

Existing tests that expect a notch to take the focus get updated to the new rule.
