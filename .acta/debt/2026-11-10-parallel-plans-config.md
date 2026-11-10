---
id: DBT-0162
hash: orcliu6
parent: plans/2026-11-10-parallel-plans-config
---
# Review NOTEs: Parallel plans: user config and CLI

- [ ] (medium) On a first run setup asks every setting acta config show lists, so it may ask about parallel_plans, subagent_picks and subagent_offer; plugin/skills/setup/SKILL.md should name these three as never asked by setup.
- [ ] (medium) acta state set exits 2 when the plan file holds uncommitted ticks from the current wave, and exits 2 again when the same text is set twice (git commit has nothing to commit); build's task loop runs state set after every task, so a wave of several tasks always hits it.
