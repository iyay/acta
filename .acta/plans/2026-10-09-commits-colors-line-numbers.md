---
parent: specs/2026-10-09-commits-colors-line-numbers
depth: minimal
id: PLN-0131
created: "2026-10-09 22:30:13"
hash: kajc0n2
started: "2026-10-09 22:31:47"
finished: "2026-10-09 22:35:11"
---
# Commit colors, diff line numbers and detail indent

**Goal:** The Commits list paints sha, task label and chore rows like lazygit, the diff box shows old and new line numbers, and the detail box header, work lines, COMMITS section and date line start in the same column as the markdown body.

**Spec:** `.acta/specs/2026-10-09-commits-colors-line-numbers.md`

**Tests:** fast `scripts/test ./internal/commits ./internal/tui`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Colors come from theme slots through `internal/tui/styles.go`, never hard-coded hex. Green = `styles.done`, red = `styles.problem`, dim = `styles.faint`. Blue is `slot(slotBlue)`; add one `blue` brush to `styles` only if none exists.
- Every drawn line stays within its width, wide runes included (wiki: xansi-wrap-overflows).
- Tests never run `rm -rf` or `rm -f`; git repos go under `t.TempDir()`.
- No version bump.

## Waves

- Wave 1: Task 01, Task 03, Task 04
- Wave 2: Task 02

### Task 01: Landed flag on commits

**Files:**
- Modify: `internal/commits/commits.go`
- Test: `internal/commits/commits_test.go`

**verify:** `Landed` is true exactly when the commit was first found from the first ref passed to `Find`, on every path. List cases checked: commit only on first ref, only on a later ref, on both refs, first ref unresolvable and skipped.

- [x] **Failing test:** temp repo with a commit on HEAD and one only on branch `feat`; `Find(repo, []Ref{{Name:"HEAD"},{Name:"feat"}})` gives `Landed` true for the HEAD one and false for the `feat` one; fails because the field does not exist.
- [x] **Change:** add `Landed bool // found from the first ref, the checked-out branch` to `Commit`; set it in `Find` from the index of the ref in the loop (index 0). A skipped first ref means no commit is landed.
- [x] **Commit:** `feat(commits): mark commits found on the checked-out branch`

### Task 02: Commit list colors

**Files:**
- Modify: `internal/tui/commits.go`, `internal/tui/styles.go`
- Test: `internal/tui/commits_test.go`

**Interfaces:**
- Consumes: `commits.Commit.Landed bool` from Task 01.

**verify:** Every list row paints its parts with one rule, and no row is wider than the box. List cases checked: landed sha green, unlanded sha red, `#n` and `#n,#m` labels blue, subject plain, chore row dim end to end, picked row keeps the full selected band with no inner color resets breaking it, row cut short mid-sha and mid-label.

- [x] **Failing test:** build rows with landed, unlanded, chore and two-label commits, render `commitListBox`, and check each part carries the expected style (compare against `styles.done/problem/faint` and the blue brush rendered output); fails because rows are plain.
- [x] **Change:** paint parts of `commitRow.text` in `commitListBox` for unpicked rows after cutting to width; picked rows keep `styles.selected` over the plain text.
- [x] **Commit:** `feat(tui): color commit list rows like lazygit`

### Task 03: Diff line numbers

**Files:**
- Modify: `internal/tui/diffcolor.go`
- Test: `internal/tui/diffcolor_test.go`

**verify:** Every line inside a hunk shows the right old and new number in a dim gutter, every other line gets a blank gutter of the same width, and no line passes width `w`. List cases checked: context line, `-` line, `+` line, `\ No newline at end of file`, two hunks, two files, `@@ -0,0 +1,3 @@` new file, `@@ -5 +5 @@` with no counts, numbers growing past the gutter width of the first hunk, stat and commit header lines.

- [x] **Failing test:** `colorDiff` over a diff with `@@ -10,3 +10,4 @@` (one context, one `-`, two `+`, one context) gives gutters `10 10`, `11   `, `   11`, `   12`, `12 13` (right-aligned, width from the largest number in the diff), header lines blank gutter; fails because there is no gutter.
- [x] **Change:** first pass reads every `@@` header to find the widest number; second pass tracks old and new counters per hunk and prefixes each line with a `styles.faint` gutter before painting; the text width left for `paintDiffLine` is `w` minus the gutter.
- [x] **Commit:** `feat(tui): show old and new line numbers in the diff`

### Task 04: Detail indent

**Files:**
- Modify: `internal/tui/detail.go`
- Test: `internal/tui/detail_test.go`, `internal/tui/commits_detail_test.go`

**verify:** In the detail box every non-markdown text line starts 2 columns in, the same column as the markdown body, and none is wider than the box. List cases checked: header field lines, `!` problem lines, work lines of a plan, task, spec, debt file and idea, COMMITS label and rows, `+n more` line, `no linked commits`, date footer long and short form, debt note middle, empty-board and no-item messages. The `───` rules stay full width.

- [x] **Failing test:** render the detail of a plan with commits and a markdown body; the first non-space column of a header line, of `COMMITS` and of a commit row equals the first non-space column of the body paragraph (2); the rule line still starts at column 0; fails because they start at 0.
- [x] **Change:** build those lines at width `w-2` and prefix two spaces in `buildDetailParts` (header, problems, work lines, commit lines) and on the date footer; leave `rule` and the glamour body as they are.
- [x] **Commit:** `feat(tui): indent the detail header and commits to the body margin`

## Polish

### Task 05: Review polish

**verify:** The NOTE below is applied, and nothing else changes.

- [x] Rewrap the `commitRowLine` doc comment in `internal/tui/commits.go` to the width of the comments around it, and rewrap the `Find` doc comment in `internal/commits/commits.go` so no line is a stub.
- [x] Commit: `polish: review notes PLN-0131`

## Review notes

- Round 1: two reviewers (full tier), both CLEAN. Polish commit is comment-only (light tier), CLEAN.
- `commitRowLine` uses a variadic identity `plainText` only to match `Style.Render`; a take that accepts a nil painter would read clearer.
- `commitRowLine` calls `r.parts` twice per visible row; cost is trivial.
- `Find` sets `Landed` before the `seen` check; harmless because seen commits are skipped right after.
- The gutter width comes from the widest number actually computed, not from `start+count-1` of each `@@` header; same result, a little more exact.
- Task 03 wrote code before its test and proved the test red by mutation afterwards.
