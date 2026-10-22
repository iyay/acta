---
id: SPC-0170
created: "2026-10-22 15:58:25"
hash: vqidxe1
started: "2026-10-22 16:00:28"
finished: "2026-10-22 16:15:34"
---
# Carry the build prompt inside each split plan

## Problem

When `plan_split` cuts a spec into plans, slice prints one short line per plan,
such as `build PLN-0160 (ready now)`. Real work often needs more than that: the
code lives in another repo, the parent branch is not the default one, landing
needs extra steps, or two sibling plans must not land at the same time. Today
that context lives only in chat. The orchestrator writes long prompts by hand
for the user to paste, and a new session that gets only `build PLN-0160`
misses it.

## Design

Bounded change approved in chat on 2026-10-22.

- `plugin/skills/slice/split.md` gets a new step: each split plan holds a
  `## Build brief` section with only the points that differ from the default:
  1. the code repo, when it is not the repo that holds the plan;
  2. the parent branch, when it is not the default branch, and a note when
     that branch is checked out nowhere;
  3. the land method, when it is not the usual `acta:land` flow;
  4. land order warnings for sibling plans that land on the same branch.
  A plan where all four are default has no section. General rules that skills
  already hold (never reset the main checkout, never push) stay out of it.
- The printed lines stay short (`build PLN-0160 (ready now)`): the plan carries
  the rest.
- `plugin/skills/build/SKILL.md`: before it picks the parent and makes the
  worktree, build reads the plan's `## Build brief`. A repo or parent branch
  named there wins over the defaults. Its land notes go on to `acta:land`.
- The brief is part of the plan, so it is written and committed on main with
  the plan. A change to it while the build runs goes in the build's worktree,
  like any plan change.

## Files and verification

Change `plugin/skills/slice/split.md`, `plugin/skills/build/SKILL.md` and
`internal/plugincheck/skill_split_test.go`.

Red first: `TestPlanSplitText` requires `## Build brief` in both skill files
and fails before the skill text changes. Run with
`scripts/test ./internal/plugincheck -run TestPlanSplitText`. Skill size caps in
`internal/plugincheck` must still pass.

No Go change in the `acta` CLI, no new eval, no change to plans that are not
split or to the land rules.
