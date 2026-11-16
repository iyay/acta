---
type: Gotcha
title: Eval graders fail good replies
description: A phrase-list regex or a haiku judge can fail a correct reply; read each red reply before calling it a skill effect
paths: [plugin/evals/]
---

A regex grader that lists phrases misses new wording. On 2026-11-16 `cold-read-planted-contradiction` failed three runs that named the conflict as "30 days vs 7 days", "can't tell which wins" or "breaks the re-download promise". The old skill text failed the same way. So a drop of one or two runs can be the grader, not the skill.

Run with `--keep-temp` and read the last reply in `out/trace.jsonl` of each red run. Widen a pattern only with that real reply as a test sample, next to a reply that must still fail. Then rerun both arms under the new pattern.

A haiku judge can vote FAIL on a reply that holds every expected quote. Count that as judge noise only after reading the reply.

An old-arm clone outside the repo needs `--trust-plugin` when no terminal is attached. Without it the run stops at once and scores nothing.
