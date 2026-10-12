---
parent: bugs/2026-10-11-tidy-names-fold-after-last-chore
id: SPC-0127
created: "2026-10-12 08:53:39"
hash: dc9ddn3
---
# acta tidy names a full fold after its newest non-chore commit

Status: Bounded, approved by the user in chat on 2026-10-12. Fixes BUG-0043.

## Problem

When no commit of a branch is kept as its own, `plan` (`internal/tidy/tidy.go:226-229`) folds the whole range into one commit and takes `src` from the newest commit of the range. `commitNode` copies only the message of `src`. When that newest commit is a planning chore, the one commit reads `chore(spec): mark started` with no body and no `Task:` trailer, while its tree holds a real `docs(wiki): ...` change.

## Design

1. In that fallback, `src` is the newest commit of the range whose subject type is not `chore`. `docs`, `wiki` and `polish` count as real here; only `chore` is passed over.
2. When every commit of the range is a `chore`, `src` stays the newest commit, as today.
3. The tree, date and `olds` of the node do not change. `withFoldedWikiChecks` still adds the `Wiki-Checked` trailers of the other commits.
4. Nothing changes when at least one commit is kept.

## Tests

- Branch with `docs(wiki): narrow paths` (body plus trailer `Task: PLN-x#1`), then `chore(plan): tick and review notes`, then `chore(spec): mark started`: tidy makes one commit whose full message is the docs commit's message, and whose tree equals the branch tree.
- `TestNoKeptCommitMakesOneCommit` and `TestWikiCheckedKeptWhenNoCommitIsKept` (all-chore ranges) pass unchanged.
- Gate: `scripts/test --full`, `go vet ./...`, `gofmt -l .` empty.

## Out of scope

- Merging `Task:` trailers from folded chore commits.
- Changing which subject types fold (`foldTypes`).
