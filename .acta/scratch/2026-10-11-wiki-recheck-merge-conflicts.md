---
id: SCR-0060
hash: mi0nngq
title: Wiki re-check timestamp bumps cause merge conflicts
status: brainstorming
created: "2026-10-11 21:04:13"
schema: "1"
started: "2026-10-11 21:04:15"
---
# Wiki re-check timestamp bumps cause merge conflicts

## Words

### 2026-10-11

Wiki re-check timestamp bumps cause merge conflicts on every cross-branch merge.

`acta wiki check <base>..<head>` flags a page stale when a file under its `paths` changed after the page's frontmatter `timestamp`. The rule says to bump `timestamp` on every re-check, even when nothing in the page was wrong. Pages with broad paths (a whole top-level folder like `tests/` or `scripts/`) go stale on almost every branch, so every branch rewrites the same `timestamp:` line with a different value and no body change. Merging two such branches conflicts on that one line, every time.

Seen in a user repo on 2026-10-11: a hotfix close merged production into development and then into staging; all six conflicts were one `timestamp:` hunk each, and one branch carried 27 timestamp-only page edits.

Goal: a "re-checked, nothing to change" verdict must not be an edit to the page file, so no timestamp-only diff exists to conflict. Optional guard against whole-top-level-folder paths. A git merge driver was set aside: it needs per-clone config and only treats the symptom.

## Context

## Log

### 2026-10-11

Round 1 answers (user: take every recommendation):
- Q1: new command `acta wiki ok <page>...` makes an empty commit with one `Wiki-Checked: <page id>` trailer per page.
- Q2: drop the `timestamp` field. Not required, ignored when present, rules stop asking for it; this repo's own pages get cleaned once.
- Q3: broad-paths guard is a rule line in build/wiki.md only, no new check severity.
- Q4: `acta tidy` carries `Wiki-Checked` trailers from folded commits onto the commit they fold into.
- Q5: one spec covers code, skill text (build/wiki.md, build/SKILL.md, land/SKILL.md) and the wiki-close eval.

### 2026-10-11

Design approved (approach 1, ancestry plus trailer; sections S1 to S5 as presented, including the new Decision page wiki-freshness-from-git).

## Open questions
