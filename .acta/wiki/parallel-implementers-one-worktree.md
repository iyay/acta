---
type: Gotcha
title: Parallel implementers in one worktree
description: Lock errors and sibling half-written files are noise; wait and retry, never reset or touch sibling files
paths: [plugin/references/house-rules.md]
timestamp: 2026-10-05T15:28:00Z
---

Two agents committing at once collide on `.git/index.lock` or ref locks. A sibling's half-written Go file breaks `go test` for the whole package, so red or green may not be yours.

Cope: on a lock error wait 15 s and retry up to 6 times. Never reset, stash, check out, or edit the sibling's file. Prove tests in a copy of HEAD plus your own files first, then re-run in the worktree once siblings commit. A test gap found after your commit is a second commit, never amend.
