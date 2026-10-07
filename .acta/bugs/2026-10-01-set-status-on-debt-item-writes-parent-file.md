---
id: BUG-0025
hash: k95mwh2
fixed_in: 45143a7
finished: "2026-10-01 16:14:10"
---
# acta set status on a debt item changes the whole debt file

## Symptom
`acta set debt/<stem>#item-1 status done` exits 0 and commits, but the item's checkbox stays open. Instead `status: done` and a `finished:` date are written into the frontmatter of the parent debt file, so the whole file reads as done. Seen by the PLN-0069 review round 1 (Standards axis) in a temp repo on 2026-10-01.

## Root cause
`SetValue` in `internal/write/ops.go` refuses `KindTask` for status ("tasks take their status from their checkboxes") but has no matching guard for `KindDebtItem`. `internal/board/board.go:105` gives a debt item the same allowed statuses as a debt file, and a debt item's `Path` is its parent debt file (`internal/board/board.go:548`), so the write lands in the parent frontmatter.

## Repro
In a temp repo, make `.acta/debt/2026-10-01-x.md` with `# Debt for PLN-1` and `- [ ] first note`, commit, then run `acta set debt/2026-10-01-x#item-1 status done`. The debt file gets `status: done`; the `- [ ]` line is unchanged.

## Found in
main (code from before PLN-0069), found by acta:review round 1 of PLN-0069.
