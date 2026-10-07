---
type: Gotcha
title: Block writes must keep symlinks
description: Temp file plus rename replaces a symlinked CLAUDE.md with a plain file; resolve the link first
paths: [internal/setup/]
timestamp: 2026-10-07T05:50:37Z
---

Many repos make CLAUDE.md a symlink to AGENTS.md. A safe write (temp file, then rename) swaps the link itself for a plain file, so the block never reaches AGENTS.md and the two files drift apart. A plain in-place write followed the link, so this broke only when the write became atomic.

Fix: `WriteBlock` runs `filepath.EvalSymlinks` first and makes the temp file in the target's folder. A dangling link is still replaced (see the PLN-0109 debt). Any new code that rewrites a user file by rename needs the same step and a symlink test.
