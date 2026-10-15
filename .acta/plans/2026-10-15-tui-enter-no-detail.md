---
title: TUI enter no longer opens the detail Implementation Plan
parent: specs/2026-10-15-tui-enter-no-detail-design
depth: minimal
id: PLN-0152
created: "2026-10-15 16:52:43"
hash: xp3dtwo
started: "2026-10-15 16:53:23"
finished: "2026-10-15 16:55:28"
---

# TUI enter no longer opens the detail Implementation Plan

**Goal:** Enter in the list panes only folds plan rows and the group row; it never moves the focus to the detail.

**Spec:** `.acta/specs/2026-10-15-tui-enter-no-detail-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Popup and form enter keys stay as they are.

## Waves

- Wave 1: Task 01

### Task 01: Enter folds only

**Files:**
- Modify: `internal/tui/model.go`, `internal/tui/hints.go`, `internal/tui/view.go`, `internal/tui/sidebar.go`
- Test: `internal/tui/*_test.go` (tests that press enter to reach the detail switch to tab)

**Seam:** `Model.Update` with a `tea.KeyMsg` for enter.

**verify:** No list row reaches the detail through enter: list every row kind (plan head, plan task, group row, plain item, item not on disk) and what enter does on it; no hint, help line or comment still says enter opens the detail.

- [x] Red test, enter keeps list focus on non-tree rows; rename `enter()` to `toggleGroup`, drop the focus move; update hints, help and comments; `go test ./internal/tui/`; commit.

## Polish

### Task 02: Review polish

**verify:** Each NOTE below applied, and nothing else changes.

- [x] hints_test failure message names shift+tab, not enter
- [x] A test checks the list hints no longer show `Detail: enter`
- [x] TestEnterNeverFocusesTheDetail checks the rows stay the same
- [x] The group row hides `Fold: enter` during a search, where enter does nothing
- [x] Commit: `polish: review notes PLN-0152`

## Review notes

- Round 1 (deep, two reviewers): CLEAN. Notes kept as notes: enter still folds the group from the detail pane (as before); the branch-item warning now shows only on e.
- Polish review (two reviewers): CLEAN. Note: the search guard on the Fold: enter hint has no test of its own.
