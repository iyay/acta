---
id: SPC-0021
hash: ythvi06
---
# Specs and Plans on Main, Worktree at Build

Status: approved by the user on 2026-09-29 (Bounded). Skill text only, no Go code.

## Why

- Today the brainstorm skill makes a worktree and commits the spec there. The spec is then hidden from other sessions on main until the branch merges, and parallel branches take the same SPEC and PLAN numbers.
- A design that gets rejected leaves a worktree with nothing to build.
- The TUI already reads every worktree and shows the copy of a plan with the most ticks (`internal/board/board.go:172`). So a plan on main with ticks in the worktree shows live progress with no code change.
- The build skill says never commit the plan file (`plugin/skills/build/SKILL.md:143`), but land needs a clean worktree before it merges (`plugin/skills/land/SKILL.md:61`). Ticks never get committed, so removing the worktree loses them.

## Design

1. `plugin/skills/brainstorm/SKILL.md`: write and commit the spec on the branch that is checked out (usually main). Drop "create its worktree now ... first commit" and "in the worktree" (lines 45, 130, 254).
2. `plugin/skills/plan/SKILL.md`: the plan is also written and committed on main. `acta:build` makes the worktree, and only after the plan is approved (line 16, next to the save line).
3. `plugin/skills/build/SKILL.md`: the worktree rule (line 34) gets one exception: the spec and plan written by brainstorm and plan stay on main. Add one line: while a build runs, do not edit its plan on main. The ticks live in the worktree and would clash at merge.
4. `plugin/skills/land/SKILL.md`: in step 2, after `acta show`, commit the plan file in the worktree as one commit `acta: tick <plan>`. Then check the worktree is clean and merge. This is the same pattern step 9 uses for debt ticks.
5. `plugin/skills/dispatch/SKILL.md` stays as it is. It gets its worktree through build.

No commit per tick: several implementers share one worktree and would fight over the git index lock, and history would fill with tick commits.

## Testing

- Grep `brainstorm` and `plan` skills: no step tells the agent to create a worktree.
- `go test ./...` stays green (some tests read skill text).
- The land eval gate runs `scripts/eval`, because the diff touches `plugin/skills/`.
