---
id: SPC-0057
created: "2026-10-01 06:16:05"
hash: w7mzlq3
---
# TUI key p sets the priority of a bug or a debt line

Status: design approved by the user in chat on 2026-10-01. Bounded: the status popup (`openPopup` in `internal/tui/model.go`), the hints (`internal/tui/hints.go`) and `acta set <id> priority` (SPC-0055, PLN-0064) already exist.

## Why

Priority can only be set from the command line today. The user wants one key in the TUI, and a hint that shows only where the key works.

## Design

- Key `p` opens a popup titled "Set priority" with the options `high`, `medium`, `low`, `none`, in that order. The cursor starts on the current value, or on `none` when the item has no priority.
- Enter writes the choice through the same path as the status popup: `m.setValue(id, "priority", value)`, so it auto-commits and shows the same status line text. Esc closes it with no write.
- `p` works only on a bug or a debt item. Any other row, an item shown from another worktree, and a legacy file are refused with a status line message and no popup, the same way `s` refuses. The message for a wrong kind is `p sets the priority of a bug or a debt line`.
- Hint `Priority: p` shows only when the selected row is a bug or a debt item that `p` would accept. It shows in the list panes and in the detail pane, next to `Status: s` or `Tick: +`.
- The help screen (`?`) gets one line: `p  set the priority of a bug or a debt line: high, medium, low or none`.

## Files

- `internal/tui/model.go`: the `p` key and the priority case in `openPopup`.
- `internal/tui/hints.go`: the hint.
- `internal/tui/view.go`: the help line.

## Testing

Every test is written first and seen failing.

- `p` on a bug and on a debt item opens the popup with the four options and the cursor on the current value (`none` when unset).
- `p` on a spec, a plan, a task and a scratch item opens no popup and sets the refusal message.
- Choosing `high` in the popup writes `priority: high` to the bug file (in a temp repo).
- The hint shows for a bug and a debt item, and not for a spec, a plan, a task or a scratch item.
- The help screen holds the new line.

## Out of scope

- A key per level (for example `1` `2` `3`).
- Setting priority on several rows at once.
