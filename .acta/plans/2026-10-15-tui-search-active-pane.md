---
parent: specs/2026-10-15-tui-search-active-pane-design
depth: minimal
id: PLN-0147
created: "2026-10-15 12:35:44"
hash: t48ia8a
---
# TUI search filters the active pane only

**Goal:** `/` filters the rows of the active pane instead of listing every item on the board.

**Spec:** `.acta/specs/2026-10-15-tui-search-active-pane-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Match rule stays as today: the trimmed, lower-cased query is inside the item's Title, Ref or Slug.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02

### Task 01: Filter the active pane's rows by the query

**Files:**
- Modify: `internal/tui/sidebar.go` (`rowsOf`, `openRows`, `activityRows`, `searchRows`)
- Test: `internal/tui/sidebar_test.go`

**Seam:** `Model.rowsOf(pane)`

**verify:** With a query set, every pane (open tab, tree tab, Activity, Done) shows only rows it would show with no query, and among those only matching rows plus the head rows (lower depth, or group rows) that have a matching row under them; no item from another tab or pane ever appears. List each pane kind checked.

- [ ] Write failing tests: two tabs each with a "foo" item, search on one tab shows only its item; Done pane with a query shows only matching done items; a plan tree with an open plan whose task matches keeps the plan head, a plan with no match is gone.
- [ ] Run `go test ./internal/tui/ -run TestSearch` and watch it fail.
- [ ] In `rowsOf`, build the normal rows, then when `m.query != ""` pass them through one filter that keeps matching rows and the heads above them. Remove the `m.query` branches from `openRows` and `activityRows`, and delete `searchRows`.
- [ ] Run `go test ./internal/tui/` and watch it pass.
- [ ] Commit.

### Task 02: Remove the unused global board search

**Files:**
- Modify: `internal/board/board.go` (`Board.Search`)
- Test: `internal/board/board_test.go` (`TestSearch`)

**Seam:** `go build ./...`

**verify:** No caller of `Board.Search` is left anywhere in the repo, and the match rule used by the TUI filter is the only search code. List the grep run.

- [ ] Delete `Board.Search` and `TestSearch`.
- [ ] Run `go build ./... && go test ./internal/board/` and watch it pass.
- [ ] Commit.
