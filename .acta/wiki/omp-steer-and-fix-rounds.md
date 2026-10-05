---
type: Gotcha
title: Steering omp mid-goal and self-started fix rounds
description: A plain prompt steers a working omp; omp starts an open fix round by itself when reply-back exits 1
paths: [plugin/omp/]
timestamp: 2026-10-06T00:00:00Z
---

A plain `herdr agent prompt <slug> "<text>"` with no `/goal` to a working omp is a steer. omp adds it to its todo list inside the running goal. Use it for a small late finding that belongs in the fix round already running; read the pane once to confirm.

When the orchestrator commits a `## Fix round` or polish task while omp still runs, omp's `acta reply-back` exits 1 on the open task and omp starts it by itself. So write the fix round section and send its `/goal` in the same turn.

Two queued `scripts/test --full` runs wait on the machine-wide `acta run-one` lock. Do not start a full suite in the main session while a dispatch runs its own. A red `internal/tui` under CPU load gets re-run alone before calling it broken.
