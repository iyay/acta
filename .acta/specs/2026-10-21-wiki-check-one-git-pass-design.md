---
parent: bugs/2026-10-21-wiki-check-slow-on-long-history
id: SPC-0169
created: "2026-10-21 09:40:36"
hash: tybscpz
---
# Wiki check reads git once per run

## Problem

`acta wiki check` runs about three full-history `git log` calls for each page:
two in `gitc.CheckCommits` and one in `gitc.Unseen`. Each call walks the whole
history, even when it finds nothing. On be-pmis (20,399 commits, 173 pages)
one call takes about 0.6 s, so the check takes about 5 minutes.

## Design

Bounded fix approved in chat on 2026-10-21.

- `wiki.Check` reads git a fixed number of times per run, whatever the number
  of pages:
  - `git rev-list --parents HEAD` for the commit graph.
  - One `log -p -U0` over the checked file, parsed into a map from page id to
    check commits. The parse rules stay as they are now: a removed line is no
    check, and a line removed and added back the same is a move, not a check.
  - One `log --full-history --no-merges --name-only` over the wiki folder and
    the union of every page's usable paths, giving the files of each commit,
    newest first.
  - One `git status` over the wiki folder for pages with uncommitted edits.
- Per page, in memory: its check commits are the commits that changed its file
  plus its checks from the map. The page is stale when the newest commit that
  touches one of its paths (matched with `covers`, the same rule `Match` uses)
  is not an ancestor of any check commit. A check commit counts as its own
  ancestor, the same as `--not` today.
- Messages, their order and what counts as stale stay exactly as now.
- `CheckCommits` and `Unseen` go away when nothing else calls them.

## Testing

- The existing `internal/wiki` and `internal/gitc` tests stay green. They hold
  what stale means: merges, moved lines, side branches.
- New red test: the number of git calls in one `Check` run does not grow with
  the number of pages.
- Measure `time acta wiki check` on be-pmis after the change.

## Out of scope

- `acta wiki ok`, `acta wiki match` and the `.checked` file format.
