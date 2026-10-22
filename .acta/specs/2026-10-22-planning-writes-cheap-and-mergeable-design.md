---
parent: scratch/2026-10-22-planning-writes-cheap-and-mergeable
id: SPC-0174
created: "2026-10-22 23:35:15"
hash: ehsbn1r
started: "2026-10-22 23:39:21"
finished: "2026-10-23 00:20:28"
---

# Cheap plan updates and conflict-free planning merges

## Goal

Agents change plans through acta commands, so they never read a whole plan to
add a fix round, a polish task, a review note or an `after:` list. Merging a
build no longer conflicts on planning files when only a date changed.

## Why

Ticks and State already go through `acta tick` and `acta state set`. Five plan
changes are still made by hand with Read and Edit. The read costs about 700 to
1,700 tokens per plan, and an agent can pay it many times in one build:

- `## Fix round <n>` from `acta:review` and `build/dispatch.md`
- the `## Polish` task
- the `## Review notes` line
- an unasked `## Verification evidence` section that agents add on their own
- the `after:` list, which `acta set` cannot change

The merge conflict was reproduced in a scratch clone. `SetField`
(`internal/write/frontmatter.go`) adds a new key at the end of the frontmatter.
A tick in a worktree adds `started:` to the spec. At the same time, `acta set`
on main adds `status:` to the same spec. Both sides insert between `hash:` and
`---`, so git marks a conflict even though the keys differ.

The plan file and the spec file have different owners. During a build, only the
worktree changes the plan, and the build skill already forbids editing it on
main. The spec is changed on main too: its status, approval edits, and the dates
`acta spec-dates` writes for other plans. Only the spec dates have to move.

## Plan 1: spec dates are written at land

- `TaskDates` (`internal/write/mark.go`) never stamps the spec when tick runs in
  a linked worktree, whatever the number of plans. Plan dates stay as they are.
  In the main checkout, tick still stamps the spec, since no land follows.
- `SpecDates` (`internal/write/specdates.go`) also writes dates for a spec with
  one plan. It takes `started` from the merged plan and writes `finished` once
  every plan of the spec is done, as it already does for several plans.
- The `acta:land` step that runs `acta spec-dates <plan hash>` applies to every
  spec.
- The wiki page `tick-stamps-spec` is rewritten to match.

Red test first: tick a task in a worktree, run `acta set <spec> status` on main,
then merge. The merge must be clean, and the spec must get its dates after
`acta spec-dates`.

## Plan 2: commands for the remaining plan writes

### `acta plan add <plan> <section> [--title T] [--verify V] < stdin`

Each stdin line becomes one item. A leading `- [ ] ` or `- ` on a line is
dropped, so the box is not written twice. An empty stdin is refused.

- `fix-round` adds `## Fix round <n>`, `### Task N: <title>`,
  `**verify:** <verify>`, one `- [ ]` box per line, and a last box for the
  commit with this task's trailer. `n` and `N` are counted from the plan.
  `--title` and `--verify` are required.
- `polish` adds `## Polish`, `### Task N: Review polish`, the verify line
  `every NOTE below is applied, and nothing else changes.`, one box per line,
  and the box ``Commit: `polish: review notes for <plan id>` ``. A plan that
  already has `## Polish` is refused.
- `review-notes` adds each line as `- <line>` under `## Review notes`. It adds
  to what is there and never replaces it. The section is created when it is
  missing.

New sections go before `## Review notes` and `## State` when they exist, and at
the end of the file otherwise. The command prints one line,
`plans/<stem>#task-N added`, so the agent can tick the new task without reading
the plan. It commits like `acta state set`, so the commit folds into the plan's
last commit. Any other section name is refused with the list of valid names.

### `acta set <plan> after <ids>`

The ids are split on commas or spaces and written as `after: [PLN-1, PLN-2]`.
The checks are the ones `acta ready` uses for exit 2: each id exists and is a
plan, the list does not name the plan itself, and there is no cycle. An empty
value removes the field.

### Guard for a building plan

Write commands run outside a linked worktree refuse to change a plan that is
being built in a worktree, with `<plan> is building in <worktree>; change it
there`. This covers `acta set`, `acta plan add` and `acta state set`. The board
already knows which plans are in a worktree (`internal/trees`).

### Skill text

- `review/SKILL.md`: the fix round, the polish task and the review notes use
  `acta plan add`. The polish template block is removed.
- `build/dispatch.md`: the fix round uses `acta plan add fix-round`, run with
  the dispatched worktree as the working directory, so the guard lets it
  through.
- `build/SKILL.md`: during a build, the plan changes only through `acta tick`,
  `acta state set` and `acta plan add`. No hand-written sections such as
  `## Verification evidence`. Findings that cost time go to
  `acta state set <plan> findings`.
- `slice/split.md`: a bad `after:` list is fixed with `acta set <plan> after`.
- `internal/plugincheck` fails when the review skill or `dispatch.md` stops
  naming `acta plan add`. `acta help` and the man page list the new command and
  the `after` field.

## Out of scope

- Spec edits stay with Edit, since a spec is free prose.
- A git merge driver for planning files. Use one only if conflicts on other
  keys keep coming after Plan 1.
- Changing where `SetField` inserts keys.

## Success

- The merge from the reproduction is clean.
- A full review round with one fix round, one polish task and review notes
  needs no Read of the plan file.
- `scripts/test --full` passes, and the review and build evals still pass.
