---
parent: scratch/2026-09-28-tui-top-tabs-per-kind
id: SPC-0012
hash: x5dx8jo
---
# TUI top tabs, one per kind

Status: design approved by the user in chat on 2026-09-28, section by section. Replaces the five-pane sidebar of `.acta/specs/2026-09-28-scratch-sidebar-design.md` (landed 622a74a).

## Why

The five-pane sidebar mixes two kinds in each box through inner tabs. The user finds it crowded and confusing. They want one top tab per kind, like the omp Settings screen, with only a list and a finished list inside each tab.

## Non-goals

- No change to kinds, statuses, files or the CLI. This is a TUI layout change only.
- No subtask rows anywhere in the lists. Subtasks stay in the detail pane SUBTASKS section.
- No saving of tab, cursor or expand state to disk.
- No newest-first sort, no themes, no drag-select copy.

## 1. Layout

```
 Scratches  Bugs  Debts  Specs  [Plans]  Activities
┌ List ──────────────┐┌ Detail ─────────────────────┐
│ + PLAN-17 ...  3/5 ││                              │
│ - PLAN-18 ...  1/4 ││                              │
│     ● task-2 ...   ││                              │
└────────────────────┘│                              │
┌ Done │ Dropped ────┐│                              │
│ + PLAN-12 ...      ││                              │
└────────────────────┘└──────────────────────────────┘
 1-6 tab · ←/→ tab · tab pane · [ ] done tab · space toggle
```

- A tab bar sits on the top line. The tabs, in order: Scratches, Bugs, Debts, Specs, Plans, Activities. The open tab is marked.
- The left column holds the panes of the open tab. The detail box stays on the right.
- Kind tabs (Scratches, Bugs, Debts, Specs, Plans) have two panes: List (open items, in-progress included) and Done. The Done pane keeps today's per-kind sub-tabs from `doneTab`: Specs and Plans Done/Dropped, Scratches Specced/Dropped, Bugs Fixed/Wontfix, Debts Done/Wontfix.
- Activities has one pane, List. It shows every task whose status is in progress, across all plans. It has no Done pane.
- The global Active pane, the global Done pane and the Tasks inner tab are removed.
- The TUI opens on Activities.
- `z` still expands the focused pane.

## 2. Plans tree

- The Plans List pane is a tree. Each plan row starts with `+` (collapsed) or `-` (expanded). An expanded plan shows its tasks under it, indented, every status, each with its status dot.
- The Plans Done pane uses the same tree.
- Every plan starts collapsed. Expand state lives in memory while the TUI runs.
- A plan with no tasks still shows `+`. Expanding it adds no rows.
- Task rows count as cursor rows, so `j`/`k` walk through them.
- A plan row selects the plan detail. A task row selects that task's detail.

## 3. Keys

- `1`..`6` jump to a tab. `←`/`→` move to the previous or next tab.
- `tab` / `shift+tab` cycle the panes of the open tab. `0` focuses the detail box.
- `[` / `]` switch the Done pane sub-tab, only while the Done pane has the focus.
- `space` / `enter` on a plan row toggle it. On a task row they do nothing.
- The `?` help popup lists these keys.

## 4. Cursor and detail

- The first visit to a tab selects the top row of its List pane, and the detail box shows that item.
- A later visit keeps the tab's last focused pane, cursor row and expanded plans.
- If a list shrank since (for example after a file watch reload), the cursor moves to the last row that still exists.
- An empty list shows `No items` in the detail box.

## 5. Code shape

- The `sidebar` table in `internal/tui/sidebar.go` becomes a `topTabs` table. Each entry has a name, a kind, its `done []doneTab`, and a `tree` flag set only for Plans. Activities is an entry with no kind and no done tabs.
- Per-tab state (focused pane, cursor per pane, expanded plans) lives in the Model, one slot per tab.
- The tree rows are built from the plan list and its tasks in one small function. It returns the rows to draw and what each row selects.
- `focusPane`, scroll, scrollbars and the popup stay as they are.

## 6. Testing

- Table: tab order, and the done sub-tabs per kind.
- Keys: `1`..`6` and `←`/`→` open the right tab. `tab` cycles only the open tab's panes. `[`/`]` only act on the Done pane.
- Cursor: the first visit selects the top row and the detail shows it. A return visit keeps pane, row and expand state. A shrunk list clamps the cursor.
- Tree: toggling changes the row count. A plan with no tasks adds no rows. A task row selects task detail, a plan row selects plan detail. The Done pane tree behaves the same.
- Activities: lists only in-progress tasks from every plan, and has no Done pane.
- Existing tests that pin the five-pane sidebar (`sidebar_test.go`, `view_test.go`, `model_test.go`) are rewritten for the new layout, not deleted without a replacement.
