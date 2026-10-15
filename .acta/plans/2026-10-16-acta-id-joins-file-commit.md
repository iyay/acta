---
title: acta id joins the file's birth commit Implementation Plan
parent: specs/2026-10-16-acta-id-joins-file-commit
depth: minimal
closes: [BUG-0045]
id: PLN-0154
created: "2026-10-16 08:14:35"
hash: vs6aayf
---

# acta id joins the file's birth commit Implementation Plan

**Goal:** A new file's id lands in the commit that brings the file in, named `new <stems>`, and no fold ever hides a file in its subject.

**Spec:** `.acta/specs/2026-10-16-acta-id-joins-file-commit.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Every `canFold` guard except the `chore(` subject check stays as it is: own author, one parent, no tag, not on a remote, on no other branch, held by no other worktree.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02

### Task 01: Folds name every file they add (BUG-0045)

**Files:**
- Modify: `internal/gitc/gitc.go`
- Test: `internal/gitc/gitc_test.go`

**verify:** After any fold, the subject names every file the folded commit adds: list each fold path checked (same file, same kind in root, new spec plus scratch parent into a `chore(scratch)` HEAD) and the subject each gives; a fold that adds nothing new keeps today's subject.

- [ ] Failing test: the BUG-0045 repro folds a new spec with its scratch parent into a `chore(scratch)` HEAD and expects the subject to name the spec stem; it fails because `foldArgs` keeps HEAD's scratch subject.
- [ ] Code: `foldArgs` rebuilds the subject as `chore: new <stems>` (kind kept when all one kind) when the fold adds a file HEAD's subject does not name.
- [ ] Commit: `fix(gitc): a fold names every file it adds`

### Task 02: acta id commits new files as `new`, folds into the adding commit

**Files:**
- Modify: `internal/write/ids.go`, `internal/gitc/gitc.go`
- Test: `internal/write/ids_test.go`, `internal/gitc/gitc_test.go`

**verify:** `acta id` never leaves a lone id commit for a file whose adding commit it may rewrite: list each case and its result: all files new (one file, spec plus plan) gives one `new <stems>` commit; a tracked file whose adding commit is HEAD folds into HEAD and keeps HEAD's subject even when it is not a `chore(` subject; HEAD pushed, tagged, on another branch, held by another worktree, or not the adding commit gives today's `assign short ids` commit; a mix of new and other files keeps today's subject.

- [ ] Failing test: `AssignIDs` on one new spec, then on a new spec plus plan, expects subjects `chore(spec): new <stem>` and `chore: new <stem>, <stem>`; and an id on a file added by a `spec: ...` HEAD expects no new commit; both fail on today's `assign short ids` path.
- [ ] Code: `AssignIDs` picks `new <stems>` when every written path was untracked; `canFold` accepts a non-`chore(` HEAD only when HEAD added every path and nothing else.
- [ ] Commit: `feat(id): a new file's id rides its birth commit`
