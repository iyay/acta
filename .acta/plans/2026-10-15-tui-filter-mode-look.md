---
parent: specs/2026-10-15-tui-filter-mode-look-design
depth: minimal
id: PLN-0150
created: "2026-10-15 13:28:05"
hash: dx5g52e
---
# TUI filter mode look

**Goal:** While typing a filter the box looks inactive with a dim selected row and the status bar shows a cursor; after enter the box is active again and the status bar says the filter is kept.

**Spec:** `.acta/specs/2026-10-15-tui-filter-mode-look-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Any dim color keeps a contrast ratio of 1.6 or more against its background.

## Waves

- Wave 1: Task 01

### Task 01: Filter mode look

**Files:**
- Modify: `internal/tui/styles.go` (`selectedDim`), `internal/tui/view.go` (`edge`, `statusLine`), `internal/tui/scroll.go` (`listView`)
- Test: `internal/tui/view_test.go`

**Seam:** `Model.edge(pane)`, `Model.listView`, `Model.statusLine()`

**verify:** While typing, the focused box border is the inactive brush, its selected row uses `selectedDim`, and the status bar ends with a cursor cell; after enter every one of these is back to normal and the status bar says `Filtered by "<query>" · esc clears` with no cursor; with no filter nothing changes. List each state checked.

- [ ] Failing tests for the typing state, the kept state and no filter.
- [ ] Add `selectedDim` (selection band, dim text, no bold); edge and listView use the inactive looks while `searching`; statusLine draws the two texts.
- [ ] Commit.
