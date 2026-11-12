---
id: SPC-0207
created: "2026-11-12 23:14:21"
hash: lt3qgwv
---
# Land runs the gates a diff needs, in any repo

Bounded. Worked out with the user in the terminal on 2026-11-12. The user approved the design in chat and ruled that the rules must hold in every project that uses acta, not only in this repo.

## Words

- Full suite: the repo's whole fast test run, named on a plan's `**Tests:**` line.
- Slow gate: any other check a repo runs before a merge that costs minutes or quota, such as behaviour evals, end-to-end tests or browser tests.

## Problem

A land spends most of its time in slow gates and in gates that run twice.

- The land skill names one repo's slow gate in its own text: `scripts/eval`, `plugin/skills/` and `plugin/hooks/`. A project with another slow gate gets no rule from the skill, and a project with none reads lines that do not apply to it.
- A slow gate runs whole whenever it runs. In this repo that is all 62 eval cases for a change to one skill.
- When the parent branch moved during the build, every gate runs again on the tip, slow gates too, even when the parent's new commits touch nothing a slow gate covers.
- A red case of a slow gate runs again alone up to 2 times, even when an open bug already lists it as flaky.

Measured in this repo on 2026-11-12: the full suite takes about 2.5 minutes alone, and the eval gate about 7 minutes at 8 runs at once.

## Changes

1. A plan names its slow gates. `acta:slice` adds a `**Gates:**` line to the plan header when the repo has a slow gate, one sentence per gate with four parts: the command; the paths whose change makes it run; the narrow form, when the gate's own tool can pick cases by tag, path or name; and the paths whose change forces the whole gate, because every case reads them. Slice takes these from the repo's own docs (`CLAUDE.md`, `AGENTS.md`, a README or a wiki page). A repo with no slow gate gets no line. The user's yes to the plan is the yes to its gates.
2. Land follows the plan's line. The land skill loses the text that names this repo's eval script and folders. For each gate on the `**Gates:**` line: it runs only when the diff touches a path that makes it run; it runs in the narrow form when the line gives one and no touched path forces the whole gate; it runs whole in every other case, and whenever the agent is in doubt about which paths the diff touches. The land report says for each gate which form ran and why, or that it did not run and why.
3. A plan with no `**Gates:**` line loses nothing. When the repo's docs name a slow gate and the plan has no line for it, land runs that gate whole. A missing line never skips a gate.
4. A narrower second run when the parent moved. The full suite runs again on the tip as it does today. A slow gate runs again only when the commits the parent gained touch a path that makes it run, and then in the form those paths call for. In every other case the report says the slow gate is reused and names the parent commits it looked at.
5. A known flake blocks once. A red case of a slow gate that an open bug lists under `flaky_tests` runs again alone once. Any other red case keeps its 2 reruns. The rules for recording a flake stay.
6. This repo's own gate, as data and not as skill text. `CLAUDE.md` names the eval gate with the four parts of change 1: `scripts/eval`; runs on a change under `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole on a change under `plugin/hooks/` or `plugin/references/`. Every eval case carries the name of the skill folder it guards as a tag, and a test in `internal/plugincheck` keeps the tag equal to the folder in the `evalCases` table.

## Risk this accepts

A narrow run can miss a case of another skill that breaks because of the change, since skills call each other. Three things limit it: a case that guards two skills carries both tags; a change to anything every case reads forces the whole gate; and a plan may name more tags on its `**Gates:**` line when its author knows of a neighbour. The whole gate still runs whenever a plan has no line.

## Files

- `plugin/skills/slice/SKILL.md`, `minimal.md`, `full.md` for the `**Gates:**` line.
- `plugin/skills/land/SKILL.md`, `tidy.md` for changes 2 to 5.
- `plugin/evals/land-eval-flake/` gets a plan with a `**Gates:**` line; one new case, `plugin/evals/land-gate-narrow/`.
- `CLAUDE.md`, the `tags:` line of each case under `plugin/evals/`, and one wiki page for change 6.
- `internal/plugincheck/skill_slice_test.go`, `skill_land_test.go`, `evals_test.go`, `budget_test.go`. A file that passes its byte cap gets a higher cap.

## Testing

- `internal/plugincheck` keeps the required strings, and fails when the land or slice skill names this repo's eval script or folders again.
- `land-gate-narrow`: a plan with a `**Gates:**` line and a diff in one area. The narrow command must run and the whole one must not.
- `land-eval-flake` and the other land cases must not get worse. The cases are measured once, at the close of the build, 10 runs each.

## Order

This plan is built first, after PLN-0231 lands, because it shortens the land of every plan behind it. PLN-0232 moves behind it.

## Out of scope

Cutting or selecting tests inside the full suite, the `run-one` lock, how many runs a case gets inside its own gate, and a new setting in `.acta.yaml`.
