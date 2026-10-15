---
parent: scratch/2026-10-16-acta-id-joins-file-commit
closes: [BUG-0045]
id: SPC-0142
created: "2026-10-16 08:12:14"
hash: gofxz4l
---
# acta id: a new file's id rides its birth commit

Status: Bounded, approved by the user in chat on 2026-10-16.

## Problem

Shape and slice write a file, run `acta id`, then `acta commit`. `acta id`
commits first, so the commit that brings a new spec or plan into the repo is
named `chore(spec): assign short ids`. In sales-agent, 19 of 26 such commits
were really the commit that added the file. The history reads as id noise
instead of "a new spec".

When the file was already committed by the user (for example
`spec: drop domain packs`), `canFold` refuses HEAD because its subject does not
start with `chore(`, so the id lands in a lone commit nobody can review or
revert on its own.

## Design

1. A call where every file it writes is new (untracked) commits all of them in
   one commit named `new <stems>`: `chore(spec): new <stem>` for one file,
   `chore: new <stem>, <stem>` when a spec and a plan come together. `Subject`
   picks the kind as it does today. Files are not split per file.
2. A file that is tracked, and whose adding commit is HEAD: the id folds into
   HEAD even when HEAD's subject does not start with `chore(`. Every other
   `canFold` guard still holds: HEAD is the user's own, one parent, no tag, not
   on a remote, on no other branch, held by no other worktree. HEAD's subject
   stays as it is.
3. Anything else (a pushed file, or HEAD did not add it) keeps today's own
   `assign short ids` commit. A call that mixes new files with step 3 files
   keeps today's subject too.
4. A scratch parent closed in the same call rides the same commit, as today.
5. A fold never hides a file. When a fold adds a file of another kind than
   HEAD's subject names, or a new file HEAD's subject does not name, the
   subject is built again to name every file the commit adds, for example
   `chore: new <scratch stem>, <spec stem>`. This closes BUG-0045: today
   `sameFiles` only checks that HEAD's files are in the new paths, so a new
   spec plus its scratch parent folds into a `chore(scratch)` commit whose
   subject never names the spec.

## Testing

Red-first tests in `internal/write/ids_test.go` and `internal/gitc/gitc_test.go`:
the `new` subject for one and for two new files, the fold into a non-chore
HEAD that added the file, and no fold when that HEAD is pushed, tagged, held
by another branch, or did not add the file; and the BUG-0045 repro, where the
folded subject must name the new spec.
