---
id: SPC-0154
created: "2026-10-19 07:14:04"
hash: xoxyifl
---
# A minimal plan waits for the user's yes

## Problem

Today a plan with `depth: minimal` skips the plan gate: slice saves it,
commits it and starts `acta:build` in the same turn, and build accepts it with
no yes. The user never sees the plan before code starts. The user ruled on
2026-10-19 that a minimal plan waits for an approve or go, the same as a full
plan.

## Design

Bounded change to the existing slice and build flow.

- `plugin/skills/slice/SKILL.md`: drop "A minimal plan needs no yes ...
  invoke `acta:build` in the same turn" and "(a minimal plan skips the wait)".
  A minimal plan is saved, given its id, committed, shown to the user, and
  waits for a yes, the same as a full plan. Depth changes only what the plan
  holds, never the gate.
- `plugin/skills/build/SKILL.md:10`: build refuses unless the plan is
  approved. The `depth: minimal` exception goes.
- Core rule 2 in `internal/hook/hook.go` and `plugin/hooks/default-rules.md`:
  drop "A plan with depth: minimal in its frontmatter needs no plan yes."
- What counts as the yes does not change: the user's approve or go in chat,
  or a direct ask to build that plan.

No change to the board: plan status is not tied to depth.

## Files

- `plugin/skills/slice/SKILL.md`, `plugin/skills/build/SKILL.md`
- `plugin/hooks/default-rules.md`, `internal/hook/hook.go`
- `internal/plugincheck/skill_slice_test.go`,
  `internal/plugincheck/skill_build_test.go`, `internal/hook/hook_test.go`,
  and byte caps in `internal/plugincheck/budget_test.go` if a size changes

## Testing

- The three test files require the new gate wording and reject the old
  exemption phrases ("needs no yes", "skips the wait", "in the same turn",
  "or has `depth: minimal`"). Red first, then green with
  `scripts/test ./internal/plugincheck/ ./internal/hook/`.
- Eval cases whose scaffold holds a `depth: minimal` plan and whose prompt
  asks to build it (for example `task-trailer`) still pass: a direct ask to
  build counts as the yes. They run in the land eval gate.
