---
type: Runbook
title: Never chain gates with merge
description: Run the gate, read its output, then merge in a separate call; a chain with ; merges on red
paths: [plugin/skills/land/]
---

Run the full gate first and read its output. Merge with `git merge` only in a later call, or join the two with `&&` on a command whose exit code is the gate itself.

Keep the gate's own log and true exit status: send both streams to a unique log, then read `$?`. Never pipe the suite into `tee` or a filter, which hides the status. Only a passing run replaces `acta-baseline.md` in the absolute Git common directory, by rename from a temp file; a red run, a reused gate or a write error leaves the last PASS and its log in place. Details live in `plugin/skills/land/baseline.md`.

Why: on 2026-10-05 a land chained the suite and the merge with `;`. The `grep -v ok` filter exits 0 either way, so a red suite printed FAIL and the merge still ran.
