---
type: Gotcha
title: scripts/eval runs one --case per call
description: With several --case flags only the last case runs; use one call per case, or --tag=<tag> for a group
paths: [scripts/eval, plugin/evals/]
---

`scripts/eval --case a --case b --runs 10` runs only `b` and exits 0, so a measure of four cases looked done after one (PLN-0232, 2026-11-13).

- One case: `scripts/eval --case <name>`.
- Several named cases: one call each, in a loop.
- A group: `scripts/eval --tag=<tag>`, which can be repeated (`--tag=land --tag=slice`). Write `--tag=` with the equals sign; see `eval-gate-by-tag.md`.

Check the `N case(s)` line at the end of the output against the number you asked for.
