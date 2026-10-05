---
type: Reference
title: Eval sandbox is greenfield
description: Each eval run gets a fresh home, cwd, and config, so no user CLAUDE.md, skills, or hooks can leak in
paths: [plugin/evals/]
timestamp: 2026-10-05T15:28:00Z
---

A run gets a fresh temp home, working dir, and Claude Code config. Nothing from the user setup loads. Each channel has a sentinel phrase checked with a `not_contains` grader, plus a `plugin-rules` control that proves the plugin itself loaded. The kept-temp trace (`--keep-temp`, read `out/trace.jsonl`) shows the reply word for word. Judge plugin behaviour only from such clean runs; dogfooding in this repo hides gaps behind the user's own CLAUDE.md.
