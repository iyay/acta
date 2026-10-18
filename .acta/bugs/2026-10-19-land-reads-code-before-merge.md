---
id: BUG-0047
hash: g71x5cb
priority: medium
started: "2026-10-19 06:17:58"
---
# Land report sometimes explains code read before the merge

## Symptom
Eval case `review-understanding-report` is flaky: about 1 run in 3 fails (judge FAIL x3) on main and on branches. The failing run never reads `src/retry.py` after the merge, so grader point 9 rejects it.

## Root cause
`plugin/skills/land/SKILL.md:80` says "read the final landed code (fixes, polish and the merge included)" but never says when or from where. An agent may read the file in the worktree before the merge and count that. In the failed run the only read was tool call 5, before `git merge --no-ff` at call 6; a passing run read it at call 8, after the merge at call 7.

Side finding in the same failed run: the agent ran `git commit --amend` on the merge commit, which the land skill never asks for.

## Repro
`scripts/eval --case review-understanding-report --runs 3 --keep-temp`, then list the tool calls in each `out/trace.jsonl` and compare the order of the `retry.py` read and the merge.

## Found in
main at d931db0 and branch setup-skill-choices, during the land gate of PLN-0162; acta:debug, 2026-10-19.
