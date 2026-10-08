---
type: Runbook
title: Never chain gates with merge
description: Run the gate, read its output, then merge in a separate call; a chain with ; merges on red
paths: [plugin/skills/land/]
timestamp: 2026-10-08T02:46:42Z
---

Run the full gate first and read its output. Merge with `git merge` only in a later call, or join the two with `&&` on a command whose exit code is the gate itself.

Why: on 2026-10-05 a land chained the suite and the merge with `;`. The `grep -v ok` filter exits 0 either way, so a red suite printed FAIL and the merge still ran.
