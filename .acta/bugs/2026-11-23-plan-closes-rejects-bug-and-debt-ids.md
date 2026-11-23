---
id: BUG-0102
hash: hy951bz
started: "2026-11-24 06:01:44"
fixed_in: PLN-jvp7jst
finished: "2026-11-24 06:09:01"
---
# A plan that closes a bug or a whole debt file shows a false problem

## Symptom
A plan with `closes: [BUG-0001]` or `closes: [DBT-0001]` (a whole debt file) shows "closes BUG-0001: a bug cannot be closed" in `acta show` and the TUI, though the land bookkeeping of SPC-0217 reads those ids from `closes:`.

## Root cause
`internal/board/closes.go:47-51` (`linkCloses`) accepts only scratch, spec and debt-line items as targets of `closes:`; every other kind gets the problem line.

## Repro
Add `closes: [BUG-nnnn]` with an existing bug id to any plan, then run `acta show <plan id>`.

## Found in
Branch land-bookkeeping-in-tidy, PLN-0250 review round 1 (Spec axis, Pre-existing).
