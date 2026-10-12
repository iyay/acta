---
id: SCR-0062
hash: xw2i9ub
title: task-trailer eval flakes on main (short id in a Bash call)
status: raw
created: "2026-10-12 07:32:04"
schema: "1"
---
# task-trailer eval flakes on main (short id in a Bash call)

## Words

### 2026-10-12

The `task-trailer` eval flakes on main: grader `no-trailer-wrong-form` fails about one run in two with "Bash called 1x (expected 0..0)", so one Bash call held the short-id form `Task: PLN-0001`.

Measured 2026-10-12 at the land of PLN-0135: branch 2 runs 0.75 / 50%, main (46b0168, main binary on PATH) 2 runs 0.75 / 50%. PLN-0135 did not touch plugin/skills/build, plugin/references or plugin/evals/task-trailer.

Not known yet: which Bash call carries the short id (the commit itself, or a probe like `git log --grep`). Next: re-run with `--keep-temp` and read trace.jsonl. Rule: never loosen the grader to go green (wiki eval-red-case-debug).

## Context

## Log

### 2026-10-12

2026-10-12, land of PLN-0136: branch runs 3, failed 2 (full suite run 0.50, then --runs 2 gave 1.00 and 0.50). Kept trace /private/tmp/e-YaFoaQ/out/trace.jsonl shows the commit itself carries the short id: `git commit -m "feat(shout): add the shout helper" ... Task: PLN-0001#1`. So the agent writes the wrong trailer form; the grader is right. The earlier `acta show PLN-0001 --path` probe also matches `PLN-0001` but not `Task: PLN-0001`. PLN-0136 changed implementer-prompt.md test lines only, not the trailer lines.

## Open questions
