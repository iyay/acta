---
id: SCR-0059
hash: gyzf4ge
title: Link old plan commits to their tasks without rewriting history
status: dropped
created: "2026-10-09 21:39:20"
schema: "1"
started: "2026-10-09 21:40:06"
finished: "2026-10-09 21:59:06"
---
# Link old plan commits to their tasks without rewriting history

## Words

### 2026-10-09

User ask, 2026-10-09, after SPC-0118 landed: the Commits view only shows commits with a `Task: PLN-<hash>#<n>` trailer, so every plan built before it shows `no linked commits`. The user wants the old ones to show too.

Idea raised in chat: attach the `Task:` line to old commits as git notes, so no sha changes. Guess the commit-to-task mapping from each plan's `started`/`finished` stamps and the scope words in commit subjects, show the guesses for review, then apply.

## Context

- SPC-0118 ruled out a backfill that rewrites main's history.
- `internal/commits` reads trailers from the commit body only; notes would need a second read path (`git log --notes` or `git notes show`).
- Old task commits have conventional subjects, empty bodies, no plan or task id.

## Log

### 2026-10-09

2026-10-09 probe round 1, user said ok to all:
- Q1 one-off for this repo, not a general acta command. Strongest signal (main-backup Land PLN- merges) exists only here.
- Q2 links live in one mapping file in .acta/, not git notes. Landed pushed commits are never rewritten, so their shas are safe to store.
- Q3 plan-level link by default; task-level only when the commit subject matches the task text exactly. Never invent a task number.
- Q4 confident links go in directly; ambiguous ones stay unlinked and go on one short list for the user.
- Q5 commits outside every confident signal stay "no linked commits"; no guessing from time stamps alone.

Measured after round 1 (515 non-chore commits on main):
- main-backup Land PLN- merges cover PLN-0035..PLN-0113 only. Mapping main shas back to backup commits is lossy: the 2026-10-07 rewrite folded and renamed commits. Tree match finds 264 main commits in backup, only 11 of them inside a Land merge.
- Plan time window plus files named in task text: 71 commits hit exactly one task, 64 hit one plan but several tasks, 1 hits several plans, 379 hit nothing.
- So confident links cover roughly a quarter of old commits at best.

### 2026-10-09

2026-10-09 probe round 2: user took the recommendation. Dropped.
Why: the 2026-10-07 history rewrite folded and renamed the old commits, so the signal is gone. Confident links would cover about a quarter of 515 old commits (about 70 task links, 60 plan links), for a one-off script, a mapping file, a second read path in internal/commits and plan-level rows in the Commits view. New plans get the Task trailer already, so the gap only shrinks.

## Open questions

- Notes live in `refs/notes/*` and need their own push and fetch; does that fit how the user shares the repo?
- Plans that ran in parallel overlap in time; how good can the guess be, and who confirms it?
- tidy rewrites shas at land; must it carry notes across, or do notes only ever go on landed commits?
- Is a reviewed mapping file in `.acta/` simpler than git notes?
