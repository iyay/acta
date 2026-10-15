---
parent: bugs/2026-10-15-search-filters-done-pane
depth: minimal
id: PLN-0148
created: "2026-10-15 12:59:45"
hash: zsrznpf
started: "2026-10-15 13:00:28"
finished: "2026-10-15 13:00:57"
---
# Search leaves the unfocused box alone

**Goal:** A `/` query narrows only the box that has the focus; the other box draws as if no search ran.

**Spec:** `.acta/specs/2026-10-15-tui-search-active-pane-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.

## Waves

- Wave 1: Task 01

### Task 01: Only the focused box sees the query

**Files:**
- Modify: `internal/tui/sidebar.go` (`rowsOf`)
- Test: `internal/tui/sidebar_test.go`

**Seam:** `Model.rowsOf(pane)`

**verify:** With a query set, the box without the focus gives exactly the rows it gives with no query (same rows, same folds), on every tab with two boxes and with either box focused; the focused box still filters. List each tab and focus checked.

- [x] Failing test: for every tab with a Done box, focus Open, set a query; Done rows equal its no-query rows. Then focus Done; Open rows equal its no-query rows.
- [x] In `rowsOf`, clear the query on the model copy when the pane is not `m.listPane()`, so `matchRows`, `isOpen` and the group fold all see no search.
- [x] Commit.

## Fix round 1

### Task 02: Focus change ends the search, other box draws plain

**Files:**
- Modify: `internal/tui/model.go` (`focusPane`), `internal/tui/sidebar.go` (`rowsOf`), `internal/tui/scroll.go` (`listView`)
- Test: `internal/tui/sidebar_test.go`

**Seam:** key presses and clicks through `Model.Update`, `Model.listView`

**verify:** Moving the focus between list boxes, by tab key or by click, always clears the query and stops typing, and a click selects the row drawn under the pointer; moving to the detail box keeps the query. The box without the focus draws the same lines with and without a query, fold marks included. List each focus path and tab checked.

- [x] Failing tests: tab from Open to Done clears the query; a click on a Done row during a search selects that drawn row; the other box's drawn lines are equal with and without a query on every two-box tab.
- [x] focusPane clears query and searching when the list box changes; one helper gives a box its view of the model, used by rowsOf and listView.
- [x] Commit.
