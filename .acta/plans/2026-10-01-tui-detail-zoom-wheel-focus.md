---
parent: specs/2026-10-01-tui-detail-zoom-wheel-focus-design
depth: minimal
id: PLN-0075
created: "2026-10-01 19:01:33"
hash: yrf1d7z
started: "2026-10-01 19:25:29"
finished: "2026-10-01 19:57:45"
---
# TUI Detail Zoom, Wheel Focus, Copy Toast Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `z` zooms the detail pane to the full body, a wheel notch over an unfocused pane focuses and scrolls it, and every copy toast reads `copied to clipboard`.

**Spec:** `.acta/specs/2026-10-01-tui-detail-zoom-wheel-focus-design.md`

**Tests:** fast `scripts/test ./internal/tui`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Mouse mode stays `tea.WithMouseCellMotion()` in `internal/cli/cli.go`. No hover tracking.
- The detail zoom reuses the narrow-screen path in `geometry()` (`internal/tui/frame.go`, `g.full`); no new layout code.
- `copy failed: <err>` and `nothing selected` stay exactly as they are.
- Every run step uses `scripts/test ./internal/tui -run <Name>`, never bare `go test` and never `./...`.
- Comments are plain English a 10-year-old can read. They say why, not what. Update comments that state the old rule (`toggleExpand`, the wheel branch in `mouse()`, `TestZTogglesAndFocusRestores`, `TestWheelOverAnUnfocusedPaneDoesNothing`).

## Waves

- Wave 1: Task 1, Task 2 (no shared file).
- Wave 2: Task 3 (shares `internal/tui/model.go` with Task 2).

### Task 1: `z` zooms the detail pane

**Files:**
- Modify: `internal/tui/sidebar.go` (`toggleExpand`), `internal/tui/frame.go` (`geometry`)
- Test: `internal/tui/view_test.go` (`TestZTogglesAndFocusRestores`)

**verify:** While the detail pane is zoomed, no frame at any width and height shows a cell of the left column, the detail box spans the full width, and the mouse `hit` finds only the detail pane; every way out (second `z`, `tab`, `shift+tab`, a number key, a click) returns the exact frame from before the zoom. List every exit checked.

- [x] Failing test: flip the "z on the detail box does nothing" block in `TestZTogglesAndFocusRestores` to expect `expanded == int(paneDetail)`, a View whose lines carry no left-column border, and the old frame after a second `z`; it fails because `toggleExpand` returns early on `paneDetail`.
- [x] Code: drop the early return in `toggleExpand`, and in `geometry()` set `wide` to `m.width >= 60 && m.expanded != int(paneDetail)` so the zoom takes the `g.full` path; skip `clampOff` side effects that do not apply to the detail pane only if a test shows they break.
- [x] Commit: `feat(tui): z zooms the detail pane to the full body`

### Task 2: the wheel focuses the pane under the pointer

**Files:**
- Modify: `internal/tui/model.go` (the wheel branch of `mouse()`)
- Test: `internal/tui/scroll_test.go` (`TestWheelOverAnUnfocusedPaneDoesNothing`, renamed to `TestWheelOverAnUnfocusedPaneFocusesIt`)

**verify:** A wheel notch over any pane of the open tab ends with that pane focused and scrolled by exactly one wheel step, the same as a notch on an already focused pane, and a notch over no pane (top bar, status line, empty cell) changes neither focus nor any offset. List every pane pair and every no-pane spot checked.

- [x] Failing test: rewrite `TestWheelOverAnUnfocusedPaneDoesNothing` so for every pair (focused p, pointer q) the notch leaves `m.focus == q` and `q` scrolled one step, plus a case over the top bar and the status line that changes nothing; it fails because the branch returns when `p != m.focus`.
- [x] Code: in the wheel branch of `mouse()`, when `p != m.focus` call `m.focusPane(p)` and go on to the scroll; if `focusPane` refused (p is not a pane of the tab, so `m.focus` is still not `p`), keep the old ignore-and-return.
- [x] Commit: `feat(tui): a wheel notch focuses the pane under the pointer`

### Task 3: copy toast says "copied to clipboard"

**Files:**
- Modify: `internal/tui/model.go` (`copyID`), `internal/tui/select.go` (`copyPicked`)
- Test: `internal/tui/model_test.go` (the `y` copy cases near line 2096 and 2126), `internal/tui/drag_test.go` (the status check near line 59)

**verify:** Every successful copy, by `y` or by drag-select, sets the status to exactly `copied to clipboard`, and no status or screen line after a copy shows the copied id or text; the failure and empty paths keep their old messages. List every copy path checked.

- [x] Failing test: change the `y` cases and the drag case to expect the status `copied to clipboard` exactly, and the View to contain it; they fail because the status still carries the id or the picked text.
- [x] Code: set `m.status = "copied to clipboard"` in `copyID` and in `copyPicked`, and remove the `lead` constant and the `xansi.Truncate` call it fed (drop the `xansi` import only if nothing else in `select.go` uses it).
- [x] Commit: `feat(tui): copy toast says copied to clipboard`
