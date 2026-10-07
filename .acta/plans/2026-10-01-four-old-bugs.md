---
parent: specs/2026-10-01-four-old-bugs-design
depth: minimal
id: PLN-0069
created: "2026-10-01 15:10:07"
hash: dxsng9w
started: "2026-10-01 15:15:17"
finished: "2026-10-01 15:38:54"
---
# Four Old Bugs Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Fix BUG-0001 (test temp folder race with git auto maintenance), BUG-0005 (wrong doctor hint), BUG-0009 (`acta set` title) and BUG-0010 (first reload resets scroll).

**Spec:** `.acta/specs/2026-10-01-four-old-bugs-design.md`

**Tests:** fast `scripts/test ./internal/write ./internal/doctor ./internal/tui ./scripts`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- acta's own commits in a user's repo keep git auto maintenance as it is; only `scripts/test` turns it off.
- Comments are plain English a 10-year-old can read. They say why, not what.

## Waves

- Wave 1: Task 1, Task 2, Task 3, Task 4 (no shared files).

### Task 1: scripts/test turns off git auto maintenance (BUG-0001)

**Files:**
- Modify: `scripts/test`
- Test: `scripts/test_test.go`

**verify:** Every `go test` that `scripts/test` starts, on the short path and on the `--full` path, runs with `GIT_CONFIG_COUNT=1`, `GIT_CONFIG_KEY_0=maintenance.auto` and `GIT_CONFIG_VALUE_0=false` in its environment. List both paths and how the test proves each.

- [x] Failing test: a new test in `scripts/test_test.go` checks that the script exports the three values before both `exec` lines; it fails because the script sets none of them.
- [x] Code: export the three values near the top of `scripts/test`, with a comment that says a commit starts a background git maintenance that can hold a lock in `.git/objects` while a test deletes its temp folder.
- [x] Commit: `scripts/test turns off git auto maintenance so temp folders delete cleanly (BUG-0001)`

### Task 2: doctor hint for a dead omp link is rm (BUG-0005)

**Files:**
- Modify: `internal/doctor/doctor.go`
- Test: `internal/doctor/doctor_test.go`

**verify:** No `stale-links` result ever suggests `omp plugin unlink`, and every dead link gets exactly one `rm <full path of the link>` hint. List every case checked (one dead link, two dead links, a live link, a plain folder) and the hint each gives.

- [x] Failing test: the one-link and two-link doctor tests expect `rm <dir>/pm` (and `rm <dir>/gstack; rm <dir>/pm`) and assert no hint holds `omp plugin unlink`; they fail because the hint is `omp plugin unlink <name>`.
- [x] Code: build the hint as `rm ` plus the full link path in the stale-links check.
- [x] Commit: `doctor tells you to rm a dead omp plugin link (BUG-0005)`

### Task 3: acta set title (BUG-0009)

**Files:**
- Modify: `internal/write/ops.go`
- Test: `internal/write/title_test.go` (new)

**verify:** After `acta set <id> title "<text>"` the board shows exactly `<text>` as the title, and nothing else in the file changes. List every case checked (first `# ` heading rewritten, frontmatter `title:` rewritten when there is no heading, heading added when there is neither, empty title refused, title with a line break refused) and what each does to the file.

- [x] Failing test: a table test runs the set command for each case in a temp repo and checks the file and the loaded board title; it fails because `title` is an unknown field.
- [x] Code: add a `title` case to the field switch that refuses empty or multi-line text, then rewrites the first `# ` body line, else the frontmatter `title:`, else adds `# <text>` at the top of the body; add `title` to the unknown field message.
- [x] Commit: `acta set can change an item title (BUG-0009)`

### Task 4: the first reload keeps the scroll of an untouched tab (BUG-0010)

**Files:**
- Modify: `internal/tui/model.go`
- Test: `internal/tui/reload_scroll_test.go` (new)

**verify:** A reload resets the detail offset only when the item on show really changes. List every case checked (fresh tab scrolled with keys then reloaded, fresh tab scrolled with the wheel then reloaded, moved cursor then reloaded, reload where the item on show is gone) and the offset each ends with.

- [x] Failing test: the repro from the bug on a fresh tab whose cursor was never moved, for keys and for the wheel, expects the scrolled offset after a same-board reload; it fails because `moveTo` sees `""` as a different item and sets the offset to 0.
- [x] Code: in `moveTo`, skip the detail offset reset when the stored cursor is `""`, with a comment that an unset cursor already shows row 0.
- [x] Commit: `a reload keeps the scroll of a tab whose cursor never moved (BUG-0010)`

## Fix round 1

Review round 1 (ccc03ef..45143a7): Spec CLEAN, Standards BLOCKED with two findings in `acta set title`.

### Task 5: acta set title touches only the title line of a story, bug, spec, plan or scratch file

**Files:**
- Modify: `internal/write/ops.go`
- Test: `internal/write/title_test.go`

**verify:** `acta set <id> title` either changes only the title text of the item named, keeping every other byte of the file (CRLF and LF line endings included), or refuses with bad input and leaves the file and git history untouched. List every id kind checked (story, bug, plan, scratch, debt file, debt item, task) and every line ending checked (LF, CRLF, with and without frontmatter), and what each does.

- [x] Failing test: new cases in `internal/write/title_test.go`: `title` on a debt item id (for example `debt/<stem>#item-1`) must be refused as bad input with the file and HEAD unchanged; a CRLF file with a heading, with and without frontmatter, must keep `\r\n` on the rewritten heading line. Both fail today: the debt item case rewrites the parent debt file's heading and commits, and the CRLF case drops the `\r`.
- [x] Code: in `SetValue`, refuse `title` for `KindDebtItem` the same way `KindTask` is refused for status; in the heading rewrite, stop the cut before a `\r` that sits right before the `\n`.
- [x] Commit: `acta set title refuses debt items and keeps CRLF line endings (BUG-0009 fix round 1)`
