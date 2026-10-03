---
name: lean
description: "acta: Use when writing or planning code. Make the smallest change that is still right, and never cut the checks that protect users or data."
---

# Lean Code

Lean means less code, never less care. The best code is the code nobody has to write.

## Understand first

Read the task. Read every file the change touches. Trace the real flow through the code, from where it starts to where it ends. Do all of this before you choose a way. Reading is never cut short.

## The ladder

Go down the rungs in order. The first one that works wins, so stop there.

1. Is it needed at all? If the need is only a guess, skip it and say so in one line.
2. Does this repo already have it? Find the code that does the job and call it.
3. Does the standard library cover it? Use that.
4. Does a native platform feature cover it? Pick it over a script or a package.
5. Does an installed dependency do the job? Use it. Do not add a new one for work a few lines can do.
6. Can it be one line? Then write one line.
7. If nothing above fits, write the minimum code that does the job.

## Rules

- No abstraction that has only one user.
- No config option for something that is always the same.
- Build nothing for a need that may never come.
- Remove before you add. Touch the fewest files.
- Two options of the same size: take the one that is right on edge cases.

## Bug fixes

A report names a symptom. Find the root cause. Before you edit, find every caller of the code you will change. Fix it once, at the point they all pass through. A patch on only the reported path leaves the other callers broken.

## Never cut

- checks at trust boundaries
- error handling that prevents data loss
- security
- accessibility
- anything the user asked for

## What you skipped

Say what you left out, in chat or in the plan, so the user can ask for it. Never leave it as a marker comment in a file.
