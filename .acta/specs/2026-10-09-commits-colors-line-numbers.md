---
created: "2026-10-09 22:40:00"
---
# Commit colors, diff line numbers and an even detail indent

Status: Bounded, approved by the user in chat on 2026-10-09.

## Problem

The Commits screen list is plain text, so a sha, a task label and a subject all look the same. The diff box has no line numbers, so a reader cannot tell where a change sits in its file. In the detail box, the header and the COMMITS section start at column 0 while the markdown body starts 2 columns in, so the text does not line up.

## Design

1. **Commit list colors, the way lazygit draws them.** No row numbers.
   - The sha is green when the commit was found from the checked-out branch (the first ref `RefsFor` returns), so it has landed. It is red when it was found only from another worktree's branch. `Find` already walks the checked-out branch first, so a commit on both keeps that branch. `commits.Commit` gets a `Landed bool` set from that first ref, so the TUI never has to guess the ref name.
   - A `#n` task label is blue. The subject keeps the body color.
   - A chore commit, shown when `c` is on, is dim from end to end.
   - The picked row keeps the selected band over the whole row.
2. **Line numbers in the diff box.**
   - Each line inside a hunk gets a dim gutter on its left: the old line number, then the new one, read from the `@@ -a,b +c,d @@` header.
   - A `-` line fills only the old number, a `+` line only the new one, a context line both.
   - File header lines and the `@@` line get an empty gutter of the same width, so the text stays in one column.
   - The gutter width fits the largest number in the diff.
3. **Detail box indent.** The header fields, the work lines (tasks, steps, debt lines, specs of an idea), the `!` problem lines, the COMMITS section and the date footer start 2 columns in, the same as the glamour body margin. The `───` rules stay full width. Text widths shrink by 2 so nothing overflows.

## Files

`internal/commits/commits.go`, `internal/tui/commits.go`, `internal/tui/diffcolor.go`, `internal/tui/detail.go`, and their tests.

## Tests

- A commit found from the first ref has `Landed` true; one found only from a later ref has it false.
- A list row paints a landed sha green, an unlanded sha red, a `#n` label blue, and a chore row dim.
- A diff with `@@ -10,3 +10,4 @@` numbers its context, `-` and `+` lines right, and the header lines get a blank gutter of the same width.
- In the detail box, the first text column of a header line and of the COMMITS lines is the same as the first text column of the body.
