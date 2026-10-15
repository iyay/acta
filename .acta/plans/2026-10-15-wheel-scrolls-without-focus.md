---
parent: specs/2026-10-15-wheel-scrolls-without-focus-design
depth: minimal
id: PLN-0146
created: "2026-10-15 12:34:13"
hash: s6unktz
started: "2026-10-15 12:34:49"
---
# Wheel scrolls without focus Implementation Plan

**Goal:** On the board, a wheel notch scrolls the pane under the pointer and never changes the focus.

**Spec:** .acta/specs/2026-10-15-wheel-scrolls-without-focus-design.md

**Tests:** scripts/test ./internal/tui, scripts/test --full

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Clicks and keys keep their focus behaviour.

## Waves

- Wave 1: Task 01

### Task 01: Wheel leaves the focus alone

**Files:**
- Modify: `internal/tui/model.go` (wheel case of the mouse handler)
- Test: `internal/tui` mouse or wheel test file, plus any existing test that expects a notch to take the focus

**Seam:** `Model.Update` with a `tea.MouseMsg` wheel event

**verify:** No wheel notch on the board ever changes `m.focus`, `m.last` or `m.expanded`, and every notch that lands on a drawn pane scrolls that pane; list each path checked (focused pane, other list pane, detail pane, zoomed pane, pane the tab does not draw, cell on no pane).

- [x] Red test, change, commit

## Fix round 1

### Task 02: Click after a wheel keeps the scroll

**Files:**
- Modify: `internal/tui/model.go` (click path of the mouse handler)
- Test: `internal/tui/scroll_test.go`, plus the stale "takes the focus" comments in `scroll_test.go`, `drag_test.go`, `tracewire_test.go`

**Seam:** `Model.Update` with a left click `tea.MouseMsg`

**verify:** A click on a visible row of a list box the wheel scrolled without the focus picks that row and never moves the box's offset.

- [x] Red test, fix, comment NOTEs, commit

## Polish

### Task 03: Review polish

**verify:** NOTE below applied, and nothing else changes.

- [x] The click test also checks that the clicked row is picked
- [x] Commit: `polish: review notes PLN-0146`

## Review notes

Deep review, round 1 BLOCKED (click after wheel jumped the list), round 2 CLEAN; polish was test-only (light tier, narrow test plus gofmt). Notes kept: a click on a list box border or title still pulls a wheeled list back to the cursor, and tab or resize does the same by design.
