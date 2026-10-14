---
id: SCR-0067
hash: a5tbuse
title: 'Tidy chore commits: no empty wiki re-checks, squash same-kind chores, fold messages that fit'
status: brainstorming
created: "2026-10-14 15:19:56"
schema: "1"
started: "2026-10-14 15:20:42"
finished: "2026-10-14 15:28:47"
---
# Tidy chore commits: no empty wiki re-checks, squash same-kind chores, fold messages that fit

## Words

### 2026-10-14

User asks, 2026-10-14:

1. acta often writes empty `chore(wiki)` commits. Cause: `acta wiki ok` (internal/cli/wiki.go wikiOk) makes an empty commit `chore(wiki): re-check <ids>` with `Wiki-Checked: <id>` trailers on purpose; page freshness is read from those trailers (wiki page wiki-freshness-from-git). PLN-0139 made 9 of them in one land.
2. On land in tidy mode, squash runs of back-to-back chore commits of the same kind: chore(wiki), chore(bug), chore(plan), chore(scratch), chore(spec).
3. A folded commit's message must fit the commits inside it, not a generic subject.
4. Commit messages should not carry `Wiki-Checked` lines.

Items 1 and 4 need a new home for the "page checked" record before the empty commits and trailers can go (for example a `checked:` field in the page frontmatter, or tidy moving the record into a kept commit). Freshness must keep working across branches without merge conflicts (internal/cli/wiki_test.go covers two branches re-checking the same page).

## Context

## Log

### 2026-10-14

Q1 where the page-checked record lives: chose option 3, an append-only log file .acta/wiki/.checked, one line per check, merged with .gitattributes merge=union. No empty commits, no Wiki-Checked trailers. The line is committed and tidy folds it into a work commit.

### 2026-10-14

Q2 line format: chose '<page id>' only. The check commit is the commit that added that line to .checked (found from git history of the file). No hash in the line, so tidy needs no remap; a line folded into a work commit makes that commit the check commit.

### 2026-10-14

Q3 where same-kind chores squash: chose option 2, at commit time. acta commit and the auto-commits amend HEAD when HEAD is a chore of the same kind and is not pushed. No rewrite of main at land.

### 2026-10-14

Q4 folded message: chose option 1. Subject is rebuilt from the folded commits: chore(<kind>): <verb> <names joined by comma>, max 72 chars, rest as +N more; mixed verbs grouped per verb. Body keeps one line per original subject. Extends existing gitc.CommitOrFold (today same file only) to same kind.

### 2026-10-14

Q5 migration: chose option 2. acta doctor --fix writes one .checked line per page that is fresh today, in one commit; then all Wiki-Checked trailer code goes (gitc reader, wiki ok writer, tidy withFoldedWikiChecks). Pages stale today stay stale.

### 2026-10-14

Section 1 approved: .acta/wiki/.checked holds one line per page '<id> <date>'; every acta write rewrites the whole file sorted, deduped, unknown pages dropped. Check commit = commit whose diff adds a '+<id> ' line (git log -p). .gitattributes merge=union, ensured by doctor --fix. wiki ok commits chore(wiki): re-check <ids> through CommitOrFold, no empty commit, no trailer. Never cleared at land: tidy would fold add+delete to nothing. doctor warns on dupes/orphans, --fix rewrites. Migration via doctor --fix writes lines for pages fresh today; all Wiki-Checked code removed.

### 2026-10-14

Section 2 approved: CommitOrFold folds same chore kind (not only same file) when all files sit under the planning root; existing guards kept. Subject rebuilt chore(<kind>): <verb> <names> grouped per verb with '; ', 72 char cap, +N more. Body one line per original subject. 'chore: assign short ids' gets a kind from its folder when single-folder.

## Open questions
