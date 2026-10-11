---
parent: scratch/2026-10-11-wiki-recheck-merge-conflicts
id: SPC-0124
created: "2026-10-11 21:07:09"
hash: lo9pw0v
---
# Wiki freshness from git, re-check without a page edit

Status: Architectural, design approved by the user in chat on 2026-10-11 (SCR-0060).

## Problem

`acta wiki check` calls a page stale when the newest commit under its `paths` is dated after the page's `timestamp` field. The only way to say "re-checked, nothing to change" is to bump that field. Pages with broad paths go stale on almost every branch, so every branch rewrites the same `timestamp:` line with a different value. Merging two such branches conflicts on that line every time.

Seen in a user repo: one hotfix close stopped on six conflicts, each a single `timestamp:` hunk, and one branch carried 27 timestamp-only page edits. `build/wiki.md` already forbids `index.md` and `log.md` because they clash between parallel worktrees; `timestamp` is the last such line.

Comparing dates is also wrong on its own: a long branch whose commits are dated before the page's timestamp passes, though the page never saw those changes.

## Goal

A re-check that finds nothing wrong leaves the page file alone, so no timestamp-only diff exists to conflict. Freshness comes from git history, not from a field in the page. A merge driver is out: it needs config in every clone and treats only the symptom.

## Design

### 1. Freshness rule (`internal/wiki/check.go`, `internal/gitc`)

- A **check commit** of a page is a commit that changed the page file, or a commit with a trailer `Wiki-Checked: <page id>`. The page id is the page path under the wiki folder without `.md`, as `acta wiki ls` names it.
- A page is stale when some commit reachable from HEAD changed a file under its `paths` and is not an ancestor of any of its check commits (a commit counts as its own ancestor). In git terms: `git rev-list HEAD --not <check commits> -- <paths>` is not empty.
- A page file with uncommitted changes in the working tree counts as checked, so "fix the page, run the check, then commit" keeps working.
- The same rule holds with and without a range. The range still only picks which pages get checked.
- The stale message names the newest unseen commit: `stale: <short sha> "<subject>" changed its paths after the page was last checked; fix the page, or run acta wiki ok <id> when nothing is wrong`.
- `timestamp` is no longer required and is ignored when present. The "timestamp is missing" message goes away. A page whose `timestamp` does not parse still loads.

### 2. `acta wiki ok <page id>...` (`internal/cli/wiki.go`)

- Every id must name a page that exists. One unknown id fails the command with no commit made.
- It makes one commit with subject `chore(wiki): re-check <ids>` and one `Wiki-Checked:` trailer per page. The commit reuses HEAD's tree, so staged changes stay staged and out of it.
- It prints the new commit's short sha.

### 3. `acta tidy` keeps the trailers (`internal/tidy/tidy.go`)

When tidy folds commits into one, every `Wiki-Checked:` trailer of the folded commits moves onto the message of the commit they fold into, with duplicates dropped. Today tidy keeps only the source commit's message, so a re-check would vanish at land.

### 4. Plugin text and eval

- `plugin/skills/build/wiki.md`: drop `timestamp` from the format and the example. Step 2 becomes: fix every wrong line, or run `acta wiki ok <id>` when the page is still right. Add a rule: `paths` names the narrowest folders or files the page is about; a whole top-level folder makes the page stale on nearly every branch.
- `plugin/skills/build/SKILL.md` (wiki step) and `plugin/skills/land/SKILL.md` (wiki gate): replace "bump the timestamp" with `acta wiki ok`.
- `plugin/evals/wiki-close`: the scaffold page loses `timestamp`; the `timestamp-bumped` grader is replaced by one that passes when the branch has a `Wiki-Checked:` trailer for the page or changed the page file. The prompt description follows.
- `plugin/evals/wiki-hint` scaffold loses `timestamp`.
- This repo: remove the `timestamp:` line from every page in `.acta/wiki/`, once.

No check for broad paths: once a re-check is an empty commit, broad paths cost one extra commit per branch, not a conflict, and a new blocking problem would break land gates that pass today.

### 5. New wiki page (approved)

Decision `wiki-freshness-from-git`: freshness comes from ancestry, a no-change re-check is `acta wiki ok`, and the page has no `timestamp` because that line conflicted on every cross-branch merge. Written in the build's wiki step, in the format of `plugin/skills/build/wiki.md`.

## Tests

- A page whose paths changed in a commit after the page's last edit is stale; the message names that commit and `acta wiki ok <id>`.
- After `acta wiki ok <id>` the same page is not stale; a later change under its paths makes it stale again.
- Two branches: branch A re-checks the page, branch B changes its paths; after merging B into A the page is stale (A's check does not see B's change). If B re-checked too, the merge is clean and not stale.
- A path commit dated earlier than the page's last edit but not its ancestor is stale (the date bug).
- An uncommitted edit to the page file makes it not stale.
- A page with no `timestamp` passes the field checks; a page with an old `timestamp` is not judged by it.
- `acta wiki ok` with an unknown id exits non-zero and makes no commit; with a staged file, the new commit's tree equals HEAD's tree and the file stays staged.
- Tidy folding a `chore(wiki)` re-check commit into a code commit keeps its `Wiki-Checked:` trailer on the result.
- Merging two branches that each re-checked the same page with `acta wiki ok` gives no conflict.
