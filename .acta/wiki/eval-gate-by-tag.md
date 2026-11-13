---
type: Decision
title: The eval gate runs by skill tag
description: Land runs scripts/eval --tag=<skill folder> for a touched skill; write --tag= with the equals sign
paths: [plugin/evals/, scripts/eval, plugin/skills/land/]
---

A land used to run all eval cases for any change under `plugin/skills/`, about 7 minutes at 8 runs at once, and again when the parent moved. Since PLN-0236 a plan names the gate on its `**Gates:**` line and land runs it only as far as the diff calls for. `CLAUDE.md` holds the gate's name, `eval`, and its four parts.

- Narrow run: `scripts/eval --tag=<skill folder>`, one flag per skill folder touched. Every case carries the folder name of the skill it guards as a tag; `TestEvalCasesCarryTheirSkillTag` fails when one is missing.
- Write `--tag=shape`, never `--tag shape`. The eval tool takes many values after `--tag`, so with a space it reads the next word, the plugin folder, as a second tag and prints "No eval cases found".
- Whole run: a change under `plugin/hooks/` or `plugin/references/`, a plan with no `**Gates:**` line, a line that names no run paths, a narrow run that picks no case, or any doubt about what the diff touches.
- A case that guards two skills carries both tags, such as `[land, review]`.

Second runs on a moved parent use these same rules. A forcing path triggers the gate itself; a gate with no run paths runs whole, even without a covered parent path.

Risk accepted: a narrow run can miss a case of another skill that breaks because skills call each other. A plan author who knows of such a neighbour adds its tag to the plan's line.
