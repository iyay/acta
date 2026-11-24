---
id: BUG-0048
hash: frnwaxi
fixed_in: 71a4d9af7f6d076384b78f54a586f5b00b9589de
finished: "2026-10-19 13:11:27"
---
# Debug skill run on a race does not force the interleaving

## Symptom
Eval case debug-race scores 0.00 on main (judge FAIL FAIL FAIL, 3 of 3 runs). The agent names the right cause but proves it only with repeated runs and a lock control, and proposes a regression test that runs 4 threads x 100 hits and checks the count, with no barrier.

## Root cause
plugin/skills/debug/SKILL.md holds a Race recipe (force the interleaving with a barrier, test holds workers the same way), but it sits in a list of conditional recipes and the run skips it: the trace shows two Bash calls, a repeat probe and a lock probe, then the reply, 5 turns. Why the recipe is skipped is not proven yet.

## Repro
scripts/eval --case debug-race --keep-temp, then read out/trace.jsonl in the kept temp folder.
Control: the skill before fd66190 with the same case copied in also scores 0.00 (2 of 2), so this is not a regression of fd66190; the case has been red since it was added there.

## Found in
main at fd66190, during the land eval gate of PLN-0170 (branch build-efficiency).
