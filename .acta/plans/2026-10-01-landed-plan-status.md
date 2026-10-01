---
parent: specs/2026-10-01-landed-plan-status-design
depth: minimal
id: PLN-0072
created: "2026-10-01 16:20:48"
hash: wrakcdv
---
# A plan's written approved or in-progress follows its task boxes Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A plan whose frontmatter says `approved` or `in-progress` shows the status its task boxes give, so a landed plan turns `done` on its own (BUG-0013).

**Spec:** `.acta/specs/2026-10-01-landed-plan-status-design.md`

**Tests:** `scripts/test ./internal/board`, `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Only `KindPlan` changes; specs, bugs, tasks, scratch and debt keep their current status rules.
- A written `draft`, `done` or `dropped` on a plan still wins; no new problem text is added for plans.
- Comments are plain English a 10-year-old reads back without stopping, and say why.

## Waves

- Wave 1: Task 1

### Task 1: Plan status follows its boxes when the written word is approved or in-progress

**Files:** `internal/board/board.go` (status switch in the derive loop, next to the `KindStory && it.plans > 0` case), `internal/board/board_test.go`

**verify:** For every plan, the board status equals `parentStatus` of its boxes with source `derived` whenever the written status is `approved` or `in-progress`, and equals the written word with source `frontmatter` whenever it is `draft`, `done` or `dropped`; no other kind's status or problems change. List each written value checked and the status it gave.

- [ ] Write a table test `TestPlanWrittenStatusFollowsBoxes` in `board_test.go` covering: written `approved` with 1/1 ticked gives `done`/`derived`; written `approved` with 1/2 ticked gives `in-progress`; written `in-progress` with 0/1 ticked gives `approved`; written `draft` and written `dropped` with all ticked keep their word with `frontmatter`; no plan gets a problem. It fails today because the `default:` branch keeps the written word.
- [ ] Add a `case it.Kind == KindPlan && (it.fmStatus == "approved" || it.fmStatus == "in-progress"):` that sets `parentStatus(it.Kind, it.plans, done, started, it.Total), "derived"`, with a short comment saying the written word is only the approval stamp, so the boxes decide; run `scripts/test ./internal/board` until green, plus `go vet ./internal/board && gofmt -l internal/board`.
- [ ] Commit `board: a plan's written approved or in-progress follows its task boxes (BUG-0013)`.
