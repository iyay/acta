---
depth: minimal
id: PLN-0149
created: "2026-10-15 13:00:00"
hash: feun362
started: "2026-10-15 13:03:16"
finished: "2026-10-15 13:03:16"
---
# Sort keeps cursor row Implementation Plan

**Goal:** Pressing `o` flips the sort of a list pane and the cursor stays on the same row number, so the view does not jump.

**Tests:** scripts/test ./internal/tui, scripts/test --full

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Approved in chat 2026-10-15: the cursor keeps its row number; the item under it changes.

## Waves

- Wave 1: Task 01

### Task 01: Sort keeps the cursor row

**Files:**
- Modify: `internal/tui/model.go` (`case "o"` in the key handler)
- Test: `internal/tui` key test file

**Seam:** `Model.Update` with key `o`

**verify:** After `o` on a list pane with the cursor on row 3 (Done list included), the cursor is still on row 3 and its scroll offset does not move.

- [x] Red test, change, commit

## Fix round 1

### Task 02: Sort leaves a wheeled list where it is

**Files:**
- Modify: `internal/tui/model.go` (`case "o"`)
- Test: `internal/tui/frame_test.go`, plus the stale doc comment in `internal/tui/order_test.go`

**verify:** After the wheel scrolls a list so the cursor is out of sight, `o` leaves the list's scroll offset as it was.

- [x] Red test, fix, commit
