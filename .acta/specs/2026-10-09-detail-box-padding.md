---
parent: plans/2026-10-09-commits-colors-line-numbers
id: SPC-0123
created: "2026-10-09 22:47:37"
hash: fbujn2o
started: "2026-10-09 22:50:44"
finished: "2026-10-09 22:57:35"
---
# Detail box padding, commit rows in the detail, no diff line numbers

Status: Bounded, approved by the user in chat on 2026-10-09. Follows SPC-0122 after the user saw it running.

## Problem

- The diff line numbers from SPC-0122 are not wanted after all.
- The detail text starts one column right of the `D` of the `Detail` title, so the box does not look even. The text also touches the top border.
- The COMMITS section of the detail draws its rows plain, while the Commits screen paints them.

## Design

1. **No line numbers in the diff.** Take the gutter out of `colorDiff` in `internal/tui/diffcolor.go`, so the diff box draws as it did before SPC-0122. Close the two gutter items of DBT-0107 (combined diff numbers, gutter color test).
2. **Even padding in the detail box.** Every line of the detail starts in the column of the `D` of the `Detail` title: one column in from the left wall. One column stays free before the right wall. One empty line sits under the top border and one above the bottom border.
   - This covers the header, the rules, the problem lines, the work lines, COMMITS, the markdown body, the debt note and the date line.
   - The markdown body margin goes from 2 to 0, and the box pad adds the one column, so the body lines up with the header.
   - The `───` rules sit inside the padding, not from wall to wall.
   - The scroll window and the scrollbar count only the lines between the two padding lines.
3. **Commit rows in the detail look like the Commits screen.** The COMMITS section draws each row with the same row painter the Commits screen uses (`commitRowLine`): sha green when landed, red when not, `#n` labels blue on a plan, subject plain. Chore commits stay hidden, the section still shows 5 rows at most, and `+n more · d to open` stays.

## Files

`internal/tui/diffcolor.go`, `internal/tui/detail.go`, `internal/tui/markdown.go`, `internal/tui/view.go` only if the top and bottom padding lines must be drawn there, and their tests.

## Tests

- A diff renders with no gutter: each line starts with its own first character.
- In the detail box, the first text column of a header line, a rule, `COMMITS`, a commit row, a body paragraph and the date line equals the column of the `D` in the title; no line passes the inner width minus one; the first and last inner lines are empty.
- A detail COMMITS row of a plan matches what `commitRowLine` paints for the same commit, label included.
