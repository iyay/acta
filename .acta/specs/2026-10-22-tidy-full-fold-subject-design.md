---
parent: bugs/2026-10-22-tidy-full-fold-subject-picks-polish
id: SPC-0176
created: "2026-10-22 23:55:15"
hash: kct4yzl
---
# Tidy keeps the real subject of a planning-only branch

## Why

When every commit of a branch folds, `fullFoldSource` in `internal/tidy/tidy.go`
picks the message of the one new commit. Its second rule takes the newest commit
whose type is not `chore`. `polish` and `wiki` are fold types too, so on a branch
that changes only planning files, a late `polish:` commit beats the earlier
`docs(wiki):` commit that holds the real change. PLN-0189 hit this.

## Design

`fullFoldSource` picks the source commit in this order:

1. The newest commit that touches a file outside the planning root (unchanged).
2. The newest commit whose type is not a fold type (`chore`, `polish`, `wiki`).
3. The newest commit whose type is not `chore`, so a branch of only polish and
   chore commits still takes the polish message.
4. The newest commit (unchanged).

The page `.acta/wiki/land-tidy-history.md` describes the rule. Its "else the
newest non-chore" text changes to match the new order.

## Verification

Two new cases in the table of `TestFullFoldTakesNewestNonChoreMessage` in
`internal/tidy/tidy_test.go`, red before the fix:

- Steps `docs(wiki)`, `chore(plan)`, `polish`, `chore(plan)`, all under `.acta/`:
  the tip carries the `docs(wiki)` message.
- Steps `polish`, `chore(plan)`, both under `.acta/`: the tip carries the
  `polish` message.

The existing cases and `TestFullFoldSubjectFromCodeCommit` stay green.

## Boundaries

No change to which commits fold or stay, to trailers, dates, or the review
marker. Only the message choice for a full fold changes.
