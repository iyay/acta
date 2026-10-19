---
id: SPC-0152
created: "2026-10-19 06:55:14"
hash: tssrpil
started: "2026-10-19 06:58:06"
finished: "2026-10-19 06:59:33"
---
# state-resume eval stops after the Next step

## Problem

The eval case `state-resume` fails about 1 run in 3 on main. In the failed
run (trace e-nMn6yY) the agent does the Next step right, then keeps going
through review and land. Land runs `git worktree remove --force wt`, which
deletes `wt/` before the sandbox stops it. Both file graders read
`wt/src/invoice.py`, so they throw "path does not exist" although the work
landed correctly on main. The prompt says "do the next step it names" but not
to stop there.

## Design

Bounded change to the eval case only. In `plugin/evals/state-resume/prompt.md`,
after the line that asks for the next step, add:

> Do only that step and commit it, then stop. Do not review or land; a later
> session does that.

Graders and `max_turns: 9` stay as they are: the case tests resuming from the
Next line without asking, not landing.

Out of scope: why land reached for `--force` and `rm -rf` (SCR-0076).

## Files

- `plugin/evals/state-resume/prompt.md`

## Testing

- `scripts/eval --case state-resume --runs 3` passes 3 of 3 (1 of 3 failed
  on main before).
- `scripts/test ./internal/plugincheck/` passes (eval layout).
