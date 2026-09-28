---
id: SCRATCH-7
hash: wo3i
title: Dispatch brief must load the build skill, not whitelist acta commands
status: raw
created: "2026-09-28"
---
User 2026-09-28, during the PLAN-18 dispatch: "harusnya bukan di brief gak sih? harusnya omp kan jalanin skill juga buat ngerjain tasknya".

What happened: the Active pane stayed empty while omp worked, because omp never ran `acta tick <task> --start`. The rule already lives in acta:build (plugin/skills/build/SKILL.md:138), but the brief never told omp to load the build skill, and its "ACTA CLI: only allowed ..." line listed only the after-commit tick, which overrode the skill.

Proposal:
- The brief template in plugin/skills/dispatch/SKILL.md gets a required line: SKILL: load build (acta:build; in omp the name is `build`) before the todo list, and follow it for every task.
- Briefs never whitelist acta commands. They may only forbid acta write commands the build skill does not call.
- herdr-delivery.md:151 then stops restating the --start rule.
- plugincheck guards for both.

Maybe part of spec 3 (SCRATCH-6, harness).
