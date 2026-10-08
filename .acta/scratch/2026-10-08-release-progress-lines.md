---
id: SCR-0057
hash: xgr0ilu
title: scripts/release prints a progress line per stage
status: raw
created: "2026-10-08 11:58:38"
schema: "1"
---
# scripts/release prints a progress line per stage

## Words

### 2026-10-08

scripts/release is silent between stages: the branch, tree and tag checks, the version bump, the commit and the tag print nothing when they pass. Only scripts/test --full output and the final push line show. User asked 2026-10-08 whether it has a verbose mode; it does not (sh -x works as a stopgap).

Idea: one short progress line per stage, for example `release: bump 0.1.43 -> 0.1.44`, `release: running scripts/test --full`, `release: commit and tag v0.1.44`. No flag needed unless the lines get noisy.

## Context

## Log

## Open questions
