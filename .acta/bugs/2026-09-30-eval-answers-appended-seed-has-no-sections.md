---
id: BUG-0012
hash: kyp33vj
status: wontfix
finished: "2026-09-30"
ref: BUG-0011
---
# The answers-appended eval fails because its seed item is in the old format

## Symptom
`scripts/eval` scores the `answers-appended` case 0.50 on every run. The `answer-in-body` grader does not find "only happens in Safari" in `.acta/scratch/eval-answer.md`. The `append-ran` grader passes, because the agent did call `acta scratch add`. It was red 3 of 3 runs on main at 9e4ca87 and 3 of 3 on the faster-agent-test-runs branch.

## Root cause
`plugin/evals/answers-appended/scaffold.sh` seeds `.acta/scratch/eval-answer.md` as an old item: no `schema: "1"` and no sections. The scratch skill (`plugin/skills/scratch/SKILL.md:36`) tells the agent to add the user's words with `--section words`. On an old item, `acta scratch add` with `--section` refuses on purpose (`internal/write/scratch.go:107`, spec `.acta/specs/2026-09-29-body-schema-design.md` line 55): `bad input: SCR-0001 is an old item with no sections`. The agent follows the skill, gets the refusal, will not edit the file by hand, and stops, so the answer never reaches the body. The seed was written before the body schema landed (PLN-0031, dc983b7), and nothing updated it after.

## Repro
`scripts/eval --case answers-appended --keep-temp`, then read `out/trace.jsonl` in the kept folder. The agent runs `printf ... | acta scratch add SCR-0001 --section words` and gets exit 1 with the message above.

Check: in a copy of `plugin/`, the seed was changed to the current format (what `acta scratch new` writes: `schema: "1"`, a `# title`, and `## Words` / `## Context` / `## Log` / `## Open questions` sections). With `claude plugin eval --trust-plugin ... --case answers-appended <copy>`, all 3 of 3 runs scored 1.00.

## Found in
main, found at the PLN-0042 land gate (eval 5/6), then debugged on 2026-09-30.
