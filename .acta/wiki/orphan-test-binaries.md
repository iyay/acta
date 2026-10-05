---
type: Gotcha
title: Orphan test binaries pin the CPU
description: Killing go test leaves its .test child running; pinned cores came from orphans plus unlocked runs
paths: [scripts/test]
timestamp: 2026-10-05T15:28:00Z
---

Killing `go test` leaves the `*.test` child alive with ppid 1 until it finishes or hits the go timeout. On 2026-10-01 four orphaned `tui.test` plus unlocked `-short` runs pinned all cores.

Cope: when CPU is pinned, look for `*.test` processes with ppid 1 first. Run suites through `scripts/test` (the `acta run-one` lock keeps one full run at a time); never run bare `go test`, and keep timeouts short so orphans die soon.
