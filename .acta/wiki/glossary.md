---
type: Glossary
title: acta glossary
description: "Short names for the moving parts: voices, rounds, stamps, gates, and the wiki itself"
paths: []
timestamp: 2026-10-05T15:28:00Z
---

fix round: one review cycle on a branch, written into the plan, run by the recipient.
gate: the checks a branch must pass before merge (`scripts/test --full`, evals where the plan needs them).
greenfield: an eval sandbox with a fresh home and config, carrying none of the user's own files.
land: merge a clean branch to main and clean up its worktree.
polish: the final review-notes commit on a branch, never a fix round of its own.
run-one: the machine-wide lock that lets one full test suite run at a time.
stamp: the started or finished line `acta tick` writes into the parent spec.
voice: the user setting in `~/.acta/config.yaml` (chat language, style, executor); setup writes it.
wiki: git-kept project notes in `.acta/wiki/`, one concept per file, no ids and no status.
worktree: one git checkout per branch where all code changes happen.
