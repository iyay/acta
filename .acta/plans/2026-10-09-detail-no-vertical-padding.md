---
parent: specs/2026-10-09-detail-box-padding
depth: minimal
id: PLN-0133
created: "2026-10-09 23:07:05"
hash: h97v83m
---
# Detail box drops its top and bottom padding

**Goal:** The detail box keeps one free column on the left and right, and no longer keeps an empty line at the top and bottom, because one terminal line looks twice as tall as one column is wide.

**Spec:** `.acta/specs/2026-10-09-detail-box-padding.md` (design point 2, amended in chat on 2026-10-09: top and bottom padding is 0)

**Tests:** fast `scripts/test ./internal/tui`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- The view and the mouse must count the same detail lines: scroll offset, scrollbar and wheel all agree.
- Executor: inline, as the user asked.

## Waves

- Wave 1: Task 01

### Task 01: No top and bottom padding

**Files:**
- Modify: `internal/tui/view.go`
- Test: `internal/tui/detail_padding_test.go`, `internal/tui/scroll_test.go`, `internal/tui/drag_test.go`, `internal/tui/view_test.go`

**verify:** In every detail box, the first inner line holds the first detail line and the last inner line holds the last one, the left and right pad stay one column, and scroll, scrollbar and wheel count the full inner height. List the box sizes checked.

- [ ] **Failing test:** `TestDetailBoxPadsEverySide` expects the first inner row to hold the ID line and the last inner row to hold the date line; fails because both rows are blank.
- [ ] **Change:** `detailArea` takes the pad off the width only; `paddedDetail` adds no empty top line; tests that skipped the padding rows stop skipping them.
- [ ] **Commit:** `feat(tui): drop the top and bottom padding of the detail box`
