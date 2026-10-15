---
parent: bugs/2026-10-15-search-filters-done-pane
depth: minimal
id: PLN-0148
created: "2026-10-15 12:59:45"
hash: zsrznpf
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

- [ ] Failing test: for every tab with a Done box, focus Open, set a query; Done rows equal its no-query rows. Then focus Done; Open rows equal its no-query rows.
- [ ] In `rowsOf`, clear the query on the model copy when the pane is not `m.listPane()`, so `matchRows`, `isOpen` and the group fold all see no search.
- [ ] Commit.
