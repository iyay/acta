---
title: TUI list filter by priority and status Implementation Plan
parent: specs/2026-10-15-tui-list-filter-design
depth: minimal
id: PLN-0153
created: "2026-10-15 17:18:47"
hash: lcuevco
started: "2026-10-15 17:20:35"
finished: "2026-10-15 17:28:07"
---

# TUI list filter by priority and status Implementation Plan

**Goal:** `f` cycles a priority filter and `F` cycles a status filter on the List box, per tab, shown in the box title.

**Spec:** `.acta/specs/2026-10-15-tui-list-filter-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Filters live in the Model only, per tab; nothing is written to disk.
- Search (`/`) keeps working and applies on top of the filters.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02

### Task 01: Filter state, keys and rows

**Files:**
- Create: `internal/tui/filter.go`
- Modify: `internal/tui/model.go`, `internal/tui/sidebar.go`
- Test: `internal/tui/filter_test.go`

**Seam:** `Model.Update` with a `tea.KeyMsg` for `f` or `F`, then `Model.openRows()`.

**verify:** The List box shows exactly the open items matching every active filter of the active tab: list each tab (Specs, Plans, Bugs, Debt, Scratch) and what `f` and `F` cycle there; priority "none" keeps only unset items; a plan head stays only when it or a task under it matches; `F` on Debt leaves rows unchanged and sets the status text; an empty result keeps the cursor in range; switching tab keeps each tab's own filters.

- [x] Red tests per key and tab; add per-tab filter state and the cycle lists in `filter.go`; handle `f`/`F` in `Update`; apply the filters in `openRows` after the sort; `go test ./internal/tui/ -run Filter`; commit.

### Task 02: Title, hints and help

**Files:**
- Modify: `internal/tui/view.go`, `internal/tui/hints.go`, `internal/tui/frame.go` (where the List box title is drawn)
- Test: `internal/tui/filter_test.go`, `internal/tui/hints_test.go`

**Seam:** `Model.View()` and the hints builder.

**verify:** The List box title shows every active filter and only those (`Bugs [H · fixing]`, `Bugs [H]`, plain `Bugs` at All) on every tab; an empty filtered list draws "no items match"; `f` and `F` show in the hints of tabs where they do something and in the `?` help.

- [x] Red tests on title, empty text, hints and help; draw them; `go test ./internal/tui/`; commit.

## Fix round 1

### Task 03: Keep the filter inside the Open list

**Files:**
- Modify: `internal/tui/sidebar.go`, `internal/tui/model.go`, `internal/tui/filter.go`, `internal/tui/hints.go`
- Test: `internal/tui/filter_test.go`

**Seam:** `Model.Update` with key messages, then `rowIDs`, `doneRowIDs`, `hints()`.

**verify:** A filter changes only the Open list of its tab: list every reader of the fold state (List rows, Done rows, space, h, l) and show the Done box and the saved folds are the same with the filter on as off; a legacy spec that passes the filter shows under its group row; `f` and `F` do nothing outside the Open list; the hints name `f` wherever it works.

- [x] Red tests: Done rows unchanged under F; space under F leaves openPlans as it was; legacy approved spec shows under F=approved; F=in-progress keeps alpha and drops lonely; F on the Done box leaves its cursor; Specs hints name f. Fix in `viewOf`, `setOpen`, `toggleRow`, `openRows`, `cycleFilter`, `filterHints`; `scripts/test ./internal/tui`; commit.

## State

### Next

Review round 3 CLEAN, polish committed, full suite green.
Not landed: user reviews the worktree first, then says land.

## Fix round 2

### Task 04: The group row follows the filter too

**Files:**
- Modify: `internal/tui/scroll.go`, `internal/tui/model.go`, `internal/tui/hints.go`
- Test: `internal/tui/filter_test.go`

**Seam:** `Model.Update` with key messages, then `rowText`, `hints()` and `groupOpen`.

**verify:** While a filter shows the legacy group open, no reader of the group fold disagrees with the screen: the arrow draws open, enter changes nothing saved, and no `Fold: enter` hint shows.

- [x] Red test on Specs `F F`: enter on the group row keeps groupOpen, the row draws ▾, the hints lack `Fold: enter`; add `filterOn()` beside `filtering()` in rowText, toggleGroup, onGroupRow; `scripts/test ./internal/tui`; commit.

## Polish

### Task 05: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] The onGroupRow comment names the filter beside the search
- [x] Commit: `polish: review notes for PLN-0153`

## Review notes

Round 3 CLEAN. Notes left as is: the group count shows every legacy spec while a filter hides some; a refused fold under a filter sets no status text; one hints test repeats another. Polish review: light tier, one comment line.

## User review fix

### Task 06: Drop wheel notches gathered before a filter change

**Files:**
- Modify: `internal/tui/model.go`
- Test: `internal/tui/filter_test.go`

**Seam:** `Model.Update` with wheel mouse messages, `F`, then `wheelTickMsg`.

**verify:** No wheel notch gathered on one filter scrolls the list under another filter, even when the same row stays under the cursor.

- [x] Red test: gather a notch, press F, tick; the offset must not move. Add the filter to `wheelMark`; `scripts/test ./internal/tui`; commit.
