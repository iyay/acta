---
parent: bugs/2026-10-19-tidy-folds-docs-work-into-wrong-message
id: SPC-0165
created: "2026-10-19 20:48:13"
hash: joz9li0
started: "2026-10-19 21:01:43"
finished: "2026-10-19 21:04:09"
---

# Tidy keeps docs work and carries Task trailers

## Why

BUG-0051. A branch whose real work is skill text lands as one commit with the
message of an unrelated bug-file edit, and every `Task:` trailer is lost.

## Design

Production code changes stay in `internal/tidy/tidy.go`.

1. Keep rule. `chore` and `wiki` commits always fold, as today. A `docs`
   commit folds only when it touches nothing outside the planning root
   (`touchesOutside` is false); otherwise it stays as its own commit.
2. The review marker wins. A `polish` commit or a `chore(plan): review notes`
   commit is still the marker: it and every commit after it fold into the
   last kept commit, whatever their type or files, as today. Rule 1 applies
   only before the marker.
3. Trailers move. When a commit folds into another one, each
   `Task: PLN-<hash>#<n>` line of the folded commit is added to the last
   paragraph of the kept commit's message, since the commit reader reads
   trailers only from that paragraph. When the kept message has no
   trailer paragraph, a new last paragraph holds them. A line is skipped when
   that last paragraph already has it; the same text elsewhere in the message
   does not count.
4. All-fold fallback. When no commit is kept, the subject comes from the
   newest commit that touches files outside the planning root. With none, the
   old rule stays. Every Task trailer of the range goes into the last
   paragraph of that message, by rule 3.
5. Wiki. Rewrite the "Tidy moves no trailers" line of `land-tidy-history`.

## Testing

Red tests in `internal/tidy/tidy_test.go`:

- A branch of `docs(skills)` (outside root), `docs(setup)` (outside root),
  `polish:` (outside root) and `docs(bug)` (planning root only) keeps the two
  `docs` commits; `polish` and `docs(bug)` fold into `docs(setup)`; the final
  tree equals the merge.
- `TestReviewMarkerFoldsRest` stays green unchanged.
- The Task trailer of a folded commit lands in the last paragraph of the kept
  commit, once, and the commit reader in `internal/commits` links both tasks.
- A kept message whose body mentions `Task: PLN-x#1` mid-text and has no
  trailer paragraph still gets that trailer in a new last paragraph.
- A branch where every commit folds takes its subject from the newest commit
  that touches files outside the planning root.

## Out of scope

Rewriting commit 2009408, which already landed.
