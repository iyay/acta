---
type: Runbook
title: Fix duplicate ids after land
description: Parallel branches pick the same next id; run acta id --fix-duplicates in main after each merge
paths: [internal/cli/]
timestamp: 2026-10-05T15:39:18Z
---

Branches cut before another lands pick the same next short id: id assignment takes max+1 of what the branch sees, and the merge has no conflict since file names differ.

After every land merge, run `acta id --fix-duplicates` in main (it auto-commits). Duplicate short ids make board refs ambiguous.
