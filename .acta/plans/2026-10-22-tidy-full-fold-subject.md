---
parent: bugs/2026-10-22-tidy-full-fold-subject-picks-polish
depth: minimal
id: PLN-0193
created: "2026-10-22 23:55:31"
hash: envp34u
started: "2026-10-22 23:55:45"
finished: "2026-10-22 23:56:32"
---

# Tidy keeps the real subject of a planning-only branch

**Goal:** When every commit of a branch folds, the one new commit takes the message of the newest commit whose type is not `chore`, `polish` or `wiki`, before falling back to the newest non-chore.
**Spec:** .acta/specs/2026-10-22-tidy-full-fold-subject-design.md
**Tests:** fast `scripts/test ./internal/tidy`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Only the message choice of a full fold changes. Which commits fold or stay, trailers, dates and the review marker stay as they are.
- Reuse `foldTypes` and `subjectType`; add no new list of types.

## Waves

- Wave 1: Task 1.

### Task 1: Skip fold types when naming a full fold

**Files:** Modify `internal/tidy/tidy.go`, `internal/tidy/tidy_test.go`; the close wiki step updates `.acta/wiki/land-tidy-history.md`.
**Seam:** `tidy.Run` over a temp repo, as `TestFullFoldTakesNewestNonChoreMessage` already drives it.
**verify:** For a full fold, the source commit is the newest that touches a file outside the planning root; else the newest whose type is not a fold type; else the newest non-chore; else the newest. No later `polish:` or `wiki` commit can take the subject from an earlier commit of another non-chore type. Every existing tidy test stays green. List each case run and the commit it picked.
- [x] Red: add two cases to the table of `TestFullFoldTakesNewestNonChoreMessage`: steps `docs(wiki)`, `chore(plan)`, `polish`, `chore(plan)` under `.acta/` want the `docs(wiki)` message; steps `polish`, `chore(plan)` under `.acta/` want the `polish` message. Run `scripts/test ./internal/tidy -run TestFullFold` and see the first case fail with the polish message.
- [x] Change: in `fullFoldSource`, add a loop before the non-chore loop that returns the newest commit whose `subjectType` is not in `foldTypes`; update the comment above the call in `plan`. Run `scripts/test ./internal/tidy`, `go vet ./internal/tidy` and `gofmt -l internal/tidy` and see them pass.
- [x] Commit: `fix(tidy): keep the real subject when a planning-only branch folds`

## Polish

### Task 2: Review polish

**verify:** NOTE below applied, and nothing else changes.
- [x] `land-tidy-history` names the last fallback again (else the newest commit) and stays at or under 250 words.
- [x] `TestFullFoldTakesNewestNonChoreMessage` is renamed `TestFullFoldTakesNewestRealMessage`, since the first rule is now "not a fold type".
- [x] Commit: `polish: review notes PLN-0193`

## Review notes

Deep, full tier (Go logic plus a wiki page agents read), two reviewers; round 1 CLEAN, no BLOCKER, no fact-check needed; polish (wiki step 4 back, test renamed) reviewed CLEAN by both axes. Notes left: the "polish and chores only" case guards the fallback and is not a red test; untyped subjects (WIP, fixup!, Revert) now beat a later polish in a full fold; the full-fold source also names the author; the old test name stays in specs and plans as history.
