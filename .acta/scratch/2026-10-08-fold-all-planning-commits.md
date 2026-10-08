---
id: SCR-0056
hash: dwo75ig
title: Fold every one-file planning commit, not only scratch add
status: raw
created: "2026-10-08 09:01:03"
schema: "1"
---
# Fold every one-file planning commit, not only scratch add

## Words

### 2026-10-08

Fold should apply to more than `scratch add`. Example: 83ab0a5 `chore(spec): assign short ids` (acta id) then 6c3a389 `chore(spec): split the review tier eval into two cases` (agent's own git commit), same spec file, two commits.

Rulings 2026-10-08 (user took every recommendation):
1. Every acta auto-commit that touches one planning file folds into HEAD when HEAD is an acta auto-commit of that same file (set, bug, debt, state, priority, scratch, id). Code commits (feat, fix) never fold.
2. Subjects differ: keep the first subject, add each later subject as a body line, so nothing is lost.
3. Fold only when the commit touches exactly one file; more files commit as today.
4. New `acta commit <path> -m <msg>` runs CommitOrFold; shape, slice and build tell agents to use it instead of git commit for planning files.
5. Wave tick commits (`chore(plan): tick wave N`) go through `acta commit` too.
The five guards from SPC-0110 stay.

## Context

## Log

## Open questions
