---
parent: specs/2026-10-09-detail-box-padding
depth: minimal
closes: [DBT-0107.01, DBT-0107.02]
id: PLN-0132
created: "2026-10-09 22:49:52"
hash: nxqb6zb
started: "2026-10-09 22:50:44"
---
# Detail box padding, commit rows in the detail, no diff line numbers

**Goal:** The diff box loses its line numbers, the detail box COMMITS rows look like the Commits screen rows, and every detail line starts in the column of the `D` of the `Detail` title with one free column on the right and one empty line at top and bottom.

**Spec:** `.acta/specs/2026-10-09-detail-box-padding.md`

**Tests:** fast `scripts/test ./internal/tui`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Every drawn line stays within its width, wide runes included (wiki: xansi-wrap-overflows).
- The view and the mouse must count the same detail lines: scroll offset, scrollbar and wheel all agree.
- Tests never run `rm -rf` or `rm -f`.
- No version bump.

## Waves

- Wave 1: Task 01, Task 02
- Wave 2: Task 03

### Task 01: No line numbers in the diff

**Files:**
- Modify: `internal/tui/diffcolor.go`
- Test: `internal/tui/diffcolor_test.go`

**verify:** No diff line carries a gutter in any form, and every line kind keeps the one style it had before SPC-0122. List the line kinds checked and the widths swept.

- [x] **Failing test:** a diff with a hunk renders each line starting with its own first character (`+`, `-`, space, `@`, `d`, `c`), no leading number or padding; fails because the gutter is there.
- [x] **Change:** remove the gutter code and its helpers from `colorDiff`; drop the gutter tests and the gutter-stripping helper in the test file; keep the style checks.
- [x] **Commit:** `feat(tui): drop line numbers from the diff`

### Task 02: Detail commit rows like the Commits screen

**Files:**
- Modify: `internal/tui/detail.go` (`commitLines` only)
- Test: `internal/tui/commits_detail_test.go`

**verify:** Every COMMITS row in the detail is exactly what `commitRowLine` paints for that commit at that width, on a task and on a plan, and chore commits never show. List cases checked: task, plan with a two-task commit, landed and unlanded, more than 5 commits, none, a chore commit.

- [x] **Failing test:** for a plan with landed, unlanded and a two-task commit, each COMMITS row equals `m.commitRowLine(row, true, w)`; for a task, `commitRowLine(row, false, w)`; fails because rows are plain.
- [x] **Change:** build the rows with `commitRowsOf(it, m.taskCommits)`, skip rows whose commit is a chore, take the first `commitRows`, and draw each with `m.commitRowLine(row, it.Kind == board.KindPlan, w)`; keep the `COMMITS` label, `no linked commits` and `+n more · d to open`.
- [x] **Commit:** `feat(tui): draw detail commit rows like the Commits screen`

### Task 03: Even padding in the detail box

**Files:**
- Modify: `internal/tui/detail.go`, `internal/tui/view.go`, `internal/tui/markdown.go`, `internal/tui/scroll.go` if the wheel counts detail lines there
- Test: `internal/tui/detail_test.go`, `internal/tui/commits_detail_test.go`, `internal/tui/view_test.go`, other tui tests whose expectations move

**verify:** In every detail box state, each non-empty line starts in the screen column of the `D` of the `Detail` title, no line reaches the last inner column, the first and last inner lines are empty, and scroll, scrollbar and wheel count the same lines. List states checked: plan, task, spec, bug, debt file, debt note, idea, empty board, no item, a box too short for the sticky header, a box 1 to 4 lines tall, widths 1 to 10.

- [ ] **Failing test:** draw the wide screen with a plan selected; find the column of `D` in the `Detail` title; the header line, the rule, `COMMITS`, a commit row, a body paragraph and the date line all start there; the first and last inner rows of the box are blank; fails because text starts one column further right and there is no top line.
- [ ] **Change:** pad the detail content once, where the box is drawn: one column left and right, one empty line top and bottom, with the detail built at the narrower width and the scroll math (`detailScroll`, `detailView`, wheel) using the smaller height. Remove the per-line `detailIndent` from `buildDetailParts`, draw the rules at the padded width, and make the markdown body start at the same column as the header (set the glamour document margin so it adds nothing on top of the pad; check no other pane uses that renderer margin). A box too small for the padding draws without it.
- [ ] **Commit:** `feat(tui): pad the detail box evenly on every side`
