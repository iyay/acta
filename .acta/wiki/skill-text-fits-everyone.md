---
type: Decision
title: Skill text fits every user
description: Skill prose stays plain, short, and free of repo paths and user names; prove rewrites with evals
paths: [plugin/skills/build/]
---

Every skill or text taken from another project is rewritten, not copied: short plain words, no repo-specific paths, no user names, so it fits any user and harness. Judge with greenfield evals plus word counts; prove the rewrite costs fewer tokens without quality loss.

A build skill starts its first task after setup, with no baseline suite. A saved suite record is history, not a setup gate, so a missing, old, unreadable or other-tree result never costs a run or a question. Each task still starts with its own failing test.
