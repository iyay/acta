---
name: land-parent-gate-triggers
description: A moved parent selects second-run gates by their plan metadata in full and tidy history.
tags: [land]
runs: 1
max_turns: 70
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash]
---

Resume acta:land at the second-run gate step. The branch suites and reviews
already passed. fixtures.md names twelve independent candidate checkouts,
their history modes and the parent-before-merge ref. Each candidate contains
the combined branch and parent tree. Tidy candidates already have the tidy
ref prepared. The parent has not moved to any candidate.

For each candidate, run the second-run full gate and the slow gates called
for by the land skill and that candidate's plan and README. Inspect the parent changes.
Do not repeat the first-run gates. Do not merge, move refs, write records,
clean up, or run other land steps. The full gate has a harmless script in
each fixture; the scripts record actual runs to the workspace calls.log.
Report which gates you ran or reused and why, then stop. Do not edit the
fixtures or their logs. Run pwd first; use the git in bin by its absolute
path, since a plain git may be unavailable in the sandbox.
