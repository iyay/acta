---
id: BUG-0007
hash: o2l7pal
started: "2026-09-29"
fixed_in: 6ca8050
finished: "2026-09-29"
---
# acta:brainstorm offers a herdr tab, eval second-brainstorm-choices fails when the skill loads

## Symptom
With the session rule fixed (PLAN-33), eval case `second-brainstorm-choices` still fails on some runs. The failing replies offer a herdr tab as one of the choices, which the grader rejects: "Does NOT offer a herdr tab, does not name HERDR_ENV". The case gives the agent only the Skill tool, so it can never read the environment.

## Root cause
`plugin/skills/brainstorm/SKILL.md:80` offers choice (b) "New herdr tab — offer this only when `HERDR_ENV=1` is set. With no herdr, do not mention it." The agent cannot check the env, so it offers the tab conditionally and names HERDR_ENV. Same failure mode PLAN-33 removed from core rule 8, one level up: the decision is left to an agent that has no way to make it.

## Repro
`scripts/eval --case second-brainstorm-choices` on branch `rule8` (commit 30dd490). Measured over 12 runs of the final rule 8 wording: 10 pass, 2 fail. The failures are this, not rule 8. Rule 8 now names both portable choices in every variant and the session text is silent about herdr outside herdr, so rule 8 is not the cause.

## Fix direction
The hook already decides the herdr line for the session text (`hook.Input.Herdr`). The skill needs the same shape: name the herdr tab only where the session text already said it, or drop it from the choices list and let the session text carry it. Do not change the grader or the prompt to go green.

## Found in
branch `rule8`, while verifying the PLAN-33 verify line, 2026-09-29.
