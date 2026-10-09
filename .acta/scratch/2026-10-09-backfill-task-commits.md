---
id: SCR-0059
hash: gyzf4ge
title: Link old plan commits to their tasks without rewriting history
status: brainstorming
created: "2026-10-09 21:39:20"
schema: "1"
started: "2026-10-09 21:40:06"
---
# Link old plan commits to their tasks without rewriting history

## Words

### 2026-10-09

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

## Open questions

- Notes live in `refs/notes/*` and need their own push and fetch; does that fit how the user shares the repo?
- Plans that ran in parallel overlap in time; how good can the guess be, and who confirms it?
- tidy rewrites shas at land; must it carry notes across, or do notes only ever go on landed commits?
- Is a reviewed mapping file in `.acta/` simpler than git notes?

## Context

## Log

## Open questions
