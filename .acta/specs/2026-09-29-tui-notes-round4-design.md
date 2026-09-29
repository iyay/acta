---
parent: scratch/2026-09-29-popup-dim-main-text
id: SPC-0027
created: "2026-09-29"
hash: tumjl17
---
# TUI notes round 4

Status: design approved by the user in chat on 2026-09-29. Architectural (layout change in several panes, new keys).

## Why

The user left ten notes on the TUI in SCR-0017: the popup dim, plus nine layout and key notes. Each one below is a small change. They ship in one plan, with one task per note, so each can be reviewed on its own.

## Layout

1. **Tab box (note 2).** The top tabs move out of the one bar line (`tabBar` in `internal/tui/view.go:118`) into their own box with a border, 3 rows tall. The box shows `1 Scratches  2 Bugs  3 Debts  4 Specs  5 Plans  6 Activities`. The active tab is highlighted. Keys `1`-`6` do not change.
2. **Pane title (note 3).** The title `List` (`tabsOf`, `internal/tui/sidebar.go:151`) becomes `Open`. On the Activities tab, the title is `Tasks`.
3. **Sort word (note 4).** The `oldest` or `newest` word leaves the top border (`internal/tui/frame.go:116`). It goes to the bottom right of the border, just before the count, like `newest · 3/12`. When the pane is narrow, the word is dropped first and the count stays.
4. **No Detail key (note 7).** The `[0]` mark on Detail (`internal/tui/sidebar.go:65`) goes away, and so does the `0` key. Detail is still reached with `tab` and `enter`. The help text drops `0`.
5. **Sticky Detail (note 6).** Detail gets three parts:
   - A sticky header. It holds every label field shown today (id, title, status, parent and the rest), except the dates.
   - A middle part that scrolls: problems, work lines, and the markdown body.
   - A sticky footer, one line: `created <date> · started <date> · finished <date>`. A date that is not set shows `-`.
   When the pane is too short to fit the header, the footer and 3 middle lines, the whole pane scrolls as one block, the way it does today.

## Behaviour

6. **Sort by id (note 9).** Today the sort key is the day from the file name (`internal/tui/order.go:16`). Items from one day tie and fall back to slug order, which looks random. The new key is the number in the short id (`SCR-0017` comes before `SCR-0018`). Ids are handed out in the order items are made, so this is creation order. When two items have the same number, the day from the file name breaks the tie. Items with no id go last. The `o` key still flips the order.
7. **Activities by parent (note 5).** Activities stops being a flat list. Tasks that are in progress are grouped under their parent, as a tree like the Plans pane. The parent is the plan. When the plan's own parent is a bug, the group head is the bug, one level only. Every group starts open. Only tasks that are in progress are listed.
8. **Collapse with h (note 8).** `h` on a task row closes its plan and moves the cursor to the plan row. `h` on a plan row closes that plan. `l` on a plan row opens it. `enter` keeps toggling as today. Left and right arrows keep switching tabs.
9. **Copy id (note 1).** `y` copies the short id of the selected row. On a task row it copies `PLN-n#task-m`. The copy goes out as OSC 52 first, so it works over SSH and tmux, and also through `pbcopy` when that command exists. The status line then shows `copied <id>`. When both ways fail, the status line shows the error.

## Popup dim (the first note)

10. When a popup opens, the text behind it should look dim, not only the border. The code already dims every line (`cover` in `internal/tui/view.go:579`, `dim` in `internal/tui/styles.go:55`), yet the user sees little change. Nobody has proven why. The first step of this task is `acta:debug` to find the cause. The likely fix is a darker dim color for the theme's dim slot. The task's test checks that every line behind the popup is drawn in the dim color and no other.

## Testing

Each note gets tests in `internal/tui` (`view_test.go`, `model_test.go`, `order_test.go`, `detail_test.go`), and they check what the user sees: the rendered lines, the cursor row, the sort order, the copied text. The clipboard is reached through one small function, so the test can swap it and check the id it was handed.

## Out of scope

- Adding a time to `created`.
- Tab numbers above 6.
- A collapse-all key.
