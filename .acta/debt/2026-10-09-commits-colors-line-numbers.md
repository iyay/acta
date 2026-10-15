---
id: DBT-0107
hash: se9mjqr
parent: plans/2026-10-09-commits-colors-line-numbers
---
# Review NOTEs: Commit colors, diff line numbers and detail indent

- [x] (low) Diff line numbers skip combined diffs: a merge commit's `@@@ -a -b +c @@@` hunk head does not match `hunkHead` in internal/tui/diffcolor.go, so its lines keep the previous hunk's numbers. Matters once a commit with a Task trailer is a merge.
- [x] (low) No test checks that the diff gutter is painted with styles.faint; the gutter tests strip colors, so a gutter drawn in the wrong color would pass.
- [-] (low) TestDetailTextStartsAtTheBodyMargin covers only a plan; problem lines, work lines of a task, spec, debt file and idea, the `+n more` and `no linked commits` lines, the short date form and the no-item messages have no indent test of their own.
