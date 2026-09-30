---
id: BUG-0011
hash: nqdr1tj
started: "2026-09-30"
fixed_in: b0135a1
finished: "2026-09-30"
---
# Saving an answer to an old scratch item fails when the agent names the words section

## Symptom
`acta scratch add SCR-0001 --section words < answer.md` exits 1 with "bad input: SCR-0001 is an old item with no sections" when the item has no body schema. The answer is not saved. The eval case answers-appended goes red (score 0.50, answer-in-body fails) on main b2bf0d1 and on branch build-executor-wording, three runs out of three.

## Root cause
internal/write/scratch.go:105-107 refuses every --section on an item without a schema, even `words`. plugin/skills/scratch/SKILL.md:36 calls `--section words` the default, so an agent that follows the skill passes it and hits the refusal. The same call with no --section appends fine.

## Repro
From the main checkout: `scripts/eval --case answers-appended --keep-temp`, then read out/trace.jsonl in the kept folder: the first Bash call is `echo "..." | acta scratch add SCR-0001 --section words` and the result is the refusal above.

## Found in
Branch build-executor-wording (PLN-0043), land-gate eval; reproduced on main b2bf0d1. The branch does not touch plugin/skills/scratch, cmd, internal/cli or internal/write.
