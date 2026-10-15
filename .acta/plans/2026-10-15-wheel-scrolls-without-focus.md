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
