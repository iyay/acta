---
parent: specs/2026-10-14-tidy-chore-commits-design
depth: minimal
id: PLN-0140
created: "2026-10-14 15:29:55"
hash: q59rynr
---
# Tidy chore commits Implementation Plan

**Goal:** Wiki checks live in `.acta/wiki/.checked` with no empty commits or `Wiki-Checked` trailers, and back-to-back planning chores of the same kind fold at commit time under a subject that names what they hold.

**Spec:** `.acta/specs/2026-10-14-tidy-chore-commits-design.md`

**Tests:** fast `scripts/test ./internal/wiki ./internal/gitc ./internal/tidy ./internal/cli ./internal/doctor ./internal/write`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- File path `.acta/wiki/.checked` (under the planning root), line form `<page id> <YYYY-MM-DD>`, sorted by page id, one line per page.
- `.gitattributes` line, verbatim: `.acta/wiki/.checked merge=union` (path follows the planning root).
- The string `Wiki-Checked` must not remain in any non-test Go file after this plan.
- Subject cap 72 characters; overflow names become `+N more`; verb groups joined with `; `.
- Tests never run `rm -rf` or `rm -f`; temp repos under `t.TempDir()`. Write commands auto-commit, so tests use temp repos only.
- No version bump. Not a task: running `acta doctor --fix` on this repo and rebuilding the PATH binary; both go in the landing report.

## Waves

- Wave 1: Task 01, Task 02
- Wave 2: Task 03, Task 04

### Task 01: freshness reads `.checked`; trailer code goes

**Files:** Create `internal/wiki/checked.go`, `internal/wiki/checked_test.go`; Modify `internal/wiki/check.go`, `internal/wiki/check_test.go`, `internal/gitc/gitc.go` (`CheckCommits`), `internal/tidy/tidy.go` (remove `withFoldedWikiChecks` and its call), `internal/tidy/tidy_test.go`, `.acta/wiki/wiki-freshness-from-git.md`, `.acta/wiki/land-tidy-history.md`

**verify:** A page is fresh exactly when every commit under its paths is an ancestor of a commit that changed the page file or added a `+<id> ` line to `.checked`. List every path checked: line added, line date changed, line only removed (never a check), duplicate lines after a union merge of two branches (no conflict, page fresh), line folded by tidy into a work commit (still fresh), orphan line, page file edit, uncommitted page; and that `Rewrite` output is sorted, deduped keeping the newest date, and drops unknown ids for every input order.

- [ ] **Failing test:** cases above in `checked_test.go` (parse and `Rewrite(lines, known ids, touched ids, today)`) and `check_test.go` / `tidy_test.go` (freshness through git). `scripts/test ./internal/wiki ./internal/tidy` fails.
- [ ] **Change:** `checked.go` holds read, `Rewrite` and write of the file. `gitc.CheckCommits` reads `git log -p --full-history --no-merges` of `.checked` for added lines plus page-file commits, no trailer query. Drop `withFoldedWikiChecks`. Rewrite both wiki pages for the new record and tidy no longer moving trailers.

### Task 02: folded subject builder

**Files:** Create `internal/gitc/subject.go`, `internal/gitc/subject_test.go`

**verify:** `FoldSubject(subjects []string) string` never drops a name silently and never passes 72 characters. List every path: same verb, mixed verbs grouped with `; `, unparsed subject as one name with empty verb, names exactly at the cap, one past the cap (`+N more`), a single subject (returned unchanged), subjects of different kinds (returns "" so the caller does not fold).

- [ ] **Failing test:** table test with the cases above. `scripts/test ./internal/gitc -run TestFoldSubject` fails.
- [ ] **Change:** parse `chore(<kind>): <verb> <names>` (names split on `, `), group by verb in first-seen order, cap and count overflow.

### Task 03: `wiki ok` writes `.checked`; doctor warns and migrates

**Files:** Modify `internal/cli/wiki.go`, `internal/cli/wiki_test.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`

**verify:** `acta wiki ok` never makes an empty commit or a trailer, and `doctor --fix` never turns a fresh page stale. List every path: `wiki ok` with known ids (file rewritten, one `chore(wiki): re-check <ids>` commit through `CommitOrFold`), unknown id (error, no write), two branches re-checking the same page then merging (no conflict); doctor on a missing union line, duplicate line, orphan line, a repo with only old trailers (`--fix` writes a line for each page fresh now and none for stale ones, then all pages keep their state), and a clean repo (no warning, no commit).

- [ ] **Failing test:** cases above in `wiki_test.go` and `doctor_test.go`. `scripts/test ./internal/cli ./internal/doctor` fails.
- [ ] **Change:** `wikiOk` calls `wiki.Rewrite` and `gitc.CommitOrFold` in place of `CommitEmpty`; doctor gains the `.checked` and `.gitattributes` checks and the fix. Migration reads freshness with the old trailer rule only inside the doctor fix, from git log, then never again.

### Task 04: same-kind chores fold

**Files:** Modify `internal/gitc/gitc.go` (`canFold`, `foldArgs`), `internal/gitc/gitc_test.go`, `internal/write/ids.go`, `internal/write/ids_test.go`

**verify:** A fold happens only when HEAD and the new commit are chores of the same kind, every file sits under the planning root or is `.gitattributes`, and every old guard holds; the folded subject comes from `FoldSubject`. List every path: two `chore(scratch): new` fold; `chore(bug)` after `chore(scratch)` does not; HEAD pushed, tagged, a merge, another author, or holding a file outside the root does not; same-file repeat still folds; `assign short ids` in one folder gets that kind and folds, across folders keeps no kind and does not.

- [ ] **Failing test:** cases above in `gitc_test.go` and `ids_test.go`. `scripts/test ./internal/gitc ./internal/write` fails.
- [ ] **Change:** `canFold` compares kind and checks the root instead of same file; `foldArgs` rebuilds the subject with `FoldSubject` and keeps the body lines; `ids.go` routes through a fold-aware commit with the folder kind.
