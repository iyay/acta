---
type: Decision
title: Skill escape lines cost proof
description: A skill line that lets agents do less when a step fails made them skip controls; prove skill text with evals
paths: [plugin/skills/debug/, plugin/evals/debug-independent-check/]
---

Found in PLN-0217 (2026-11-07), on the debug-independent-check eval:

- The line "No copy? Run only probes that write nothing" made agents skip the control on the real command. With it, `behavior.md` passed 3 of 15 runs; without it, 3 of 3. A fallback that allows less work reads as leave to do less, even when the normal path works.
- Cutting "(in the temp copy)" and one Phase 1 line from the debug skill to stay under its byte cap also cost controls. Raise the cap in `internal/plugincheck/budget_test.go` instead of trimming words an agent acts on.
- Agents copy a command pattern shown in the skill line (`D=$(mktemp -d) && cd "$D"`) and do not call a helper script the line names. A helper is proven only by its own test, not by the eval.
- In the eval sandbox `mktemp -d` fails with "Operation not permitted", while `mkdir -p "$TMPDIR/name"` works. In a normal session `mktemp -d` works.
- A judge grader counts a write command that never ran, for example a `sed -i` after a failed `&&` step, unless the grader says it does not count. Check the kept repo files (`--keep-temp`, then the `sealed/home/cwd` folder) before trusting a repo-write FAIL.
