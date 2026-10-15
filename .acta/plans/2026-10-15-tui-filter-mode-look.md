---
parent: specs/2026-10-15-tui-filter-mode-look-design
depth: minimal
id: PLN-0150
created: "2026-10-15 13:28:05"
hash: dx5g52e
started: "2026-10-15 13:28:06"
finished: "2026-10-15 13:30:12"
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

- [x] Failing tests for the typing state, the kept state and no filter.
- [x] Add `selectedDim` (selection band, dim text, no bold); edge and listView use the inactive looks while `searching`; statusLine draws the two texts.
- [x] Commit.

## Fix round 1

### Task 02: Dim band keeps its contrast

**Files:**
- Modify: `internal/tui/styles.go` (`selectedDim`), `internal/tui/view.go` (`filterCursor` comment place)
- Test: `internal/tui/styles_test.go`

**Seam:** `newStyles(theme, dark)`

**verify:** On every built-in theme with hex colors, the `selectedDim` text keeps a contrast of 1.6 or more over its band, and the band is calmer than the `selected` band (closer to the theme background). List each theme checked.

- [x] Failing test over every built-in theme.
- [x] The dim band is the selection band mixed half way toward the theme background, with the selection text, not bold; `statusLine` keeps its own doc comment.
- [x] Commit.

## Polish

### Task 03: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] The contrast test also checks that the dim band is never bold.
- [x] Commit: `polish: review notes for PLN-0150`

## Review notes

Deep review, two rounds plus polish: notes left as is were the cursor glyph width on CJK terminals and contrived low-contrast user themes; two medium notes went to debt.
