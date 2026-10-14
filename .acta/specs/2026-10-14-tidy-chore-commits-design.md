---
parent: scratch/2026-10-14-tidy-chore-commits
id: SPC-0130
created: "2026-10-14 15:28:47"
hash: d5cq97i
started: "2026-10-14 15:40:18"
finished: "2026-10-14 15:56:09"
---
# Tidy chore commits: wiki checks in a file, same-kind chores fold

Status: Architectural, approved by the user in chat on 2026-10-14, section by section.

## Problem

- `acta wiki ok` makes an empty `chore(wiki): re-check <ids>` commit with `Wiki-Checked: <id>` trailers. PLN-0139 made 9 of them in one land.
- `acta tidy` moves those trailers onto the commit they fold into, so kept commit messages carry `Wiki-Checked` lines.
- On main, back-to-back chores of the same kind (`chore(scratch)`, `chore(plan)`, `chore(debt)`) each stand alone. Tidy folds parent chores only when nothing but chores sits since the last land, and one `spec:` or `plan:` commit cancels that.
- When `gitc.CommitOrFold` folds, the subject stays the first commit's subject, so it does not describe what is inside.

## Goal

1. No empty commits and no `Wiki-Checked` trailers. Wiki freshness keeps working across branches with no merge conflicts.
2. Back-to-back chores of the same kind fold at commit time.
3. A folded commit's subject names what it holds.

## Design

### 1. Wiki check record in `.acta/wiki/.checked`

- One line per page: `<page id> <UTC time, RFC 3339>`, for example `wiki-freshness-from-git 2026-10-14T08:50:01Z`. The time only makes the line change on each check; a date alone would leave a second check on the same day with no diff and no commit. Freshness never reads it.
- Every acta write of the file rewrites it whole: sorted by page id, one line per page (newest date wins), lines for pages that no longer exist dropped.
- `.gitattributes` holds `.acta/wiki/.checked merge=union`, so two branches that both re-check a page merge with no conflict. The duplicate line this leaves is removed by the next rewrite.
- `acta wiki ok <id>...` rewrites the file with the current time on those pages and commits it as `chore(wiki): re-check <ids>` through `gitc.CommitOrFold`. No empty commit, no trailer.
- Check commit of a page: a commit that changed the page file, or a commit whose diff adds a line `+<id> ` to `.checked` (read from `git log -p --full-history --no-merges -- .acta/wiki/.checked`). A commit that only removes a line is never a check commit. Every other freshness rule stays as it is today (ancestry, `--full-history --no-merges`, uncommitted page counts as checked).
- The file is never cleared at land. Tidy would fold the add and the delete into one commit with no diff, and the check would be lost.
- `acta doctor` warns when `.checked` has duplicate or orphan lines, or `.gitattributes` lacks the union line. `acta doctor --fix` rewrites the file and adds the line.
- Migration: `acta doctor --fix` writes one line for every page that is fresh today, in one commit. Pages stale today stay stale.
- Remove all `Wiki-Checked` code except the one migration reader inside the doctor fix (`internal/doctor/wiki.go`): the reader in `internal/gitc/gitc.go` (around line 303), the writer in `internal/cli/wiki.go` (around line 106), `withFoldedWikiChecks` in `internal/tidy/tidy.go`, and `gitc.CommitEmpty`.
- `internal/cli/wiki_test.go` two-branch case must still pass: both branches re-check the same page, the merge has no conflict, and the page is fresh.

### 2. Same-kind chores fold at commit time

`gitc.CommitOrFold` (`internal/gitc/gitc.go`) and its `canFold` check change in one rule only.

- Old rule: HEAD holds only the same file.
- New rule: HEAD's subject has the same kind as the new message (`chore(scratch)` with `chore(scratch)`), and every file in HEAD and the new file sit under the planning root, or are `.gitattributes`.
- Guards kept: no tag on HEAD, one parent, same author, not pushed or fetchable from a remote, no other dirty change in the file.
- `chore: assign short ids` gets a kind from its folder (`chore(spec)`, `chore(plan)`, ...) when all its files sit in one folder. Files from several folders keep no kind and never fold.

### 3. Folded subject

- Parse each subject as `chore(<kind>): <verb> <names>`. The verb is the first word, names are the rest split on `, `. A subject that does not parse counts as one name with an empty verb.
- Build `chore(<kind>): <verb> <name>, <name>`. Different verbs group per verb, joined with `; `. Example: `chore(debt): tick DBT-0044.05, DBT-0034.07`.
- Cap at 72 characters. Names that do not fit become `+N more`.
- Body: one line per original subject, before any trailer block, as today.

## Wiki

- Rewrite `wiki-freshness-from-git` for the `.checked` file: what the check commit is, why one line per page, why never cleared at land, why union merge.
- Update `land-tidy-history`: tidy no longer moves `Wiki-Checked` trailers.

## Testing

- Freshness: check by `.checked` line, a removal-only commit is not a check, two-branch merge with no conflict, page fresh after tidy folds the line into a work commit.
- Rewrite: dedupe keeps newest, orphan dropped, sorted output.
- Doctor: warns on dupes, orphans, missing union line; `--fix` migrates a repo with trailers and leaves fresh pages fresh.
- Subject: table test for same verb, mixed verbs, unparsed subject, 72-char overflow.
- Fold: two `chore(scratch): new` fold into one; `chore(bug)` after `chore(scratch)` does not; a pushed HEAD does not; a commit with a file outside the planning root does not.
