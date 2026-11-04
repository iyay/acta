---
name: slice
description: "acta: Use after a design is approved, before any code. Turns the spec into a task-by-task plan in .acta/plans/ - failing test first, exact files, a property-shaped verify line and a wave for every task - then waits for the user's yes."
---

# Writing Plans

## Overview

Write the plan for a skilled implementer who knows nothing about this codebase, its tools or its domain, and who sees only their own task. Each task names the files to touch, the code, the tests and how to run them. Bite-sized tasks. DRY. YAGNI. TDD. One commit per task.

**Context:** `acta:build` makes the worktree, only after the plan is approved.

**Save plans to:** `.acta/plans/YYYY-MM-DD-<feature-name>.md` (`.acta/` is the default root; `.acta.yaml`, `ACTA_ROOT` or `acta --root` can move it).
- (User preferences for plan location override this default)
- Save, then run `acta id` right after so the plan gets its PLN number and hash before anyone cites it.
- Commit the plan on main with `acta commit <path> -m "<message>"`, never `git commit`; commit every later edit to a plan or spec the same way. The worktree branch starts from that commit, so it carries the plan.
- A fix round, or any change to a plan whose build is running, goes in that build's worktree, not on main.

## Plan depth

Pick the depth in this order, and do not ask when one answers: the argument of `/slice minimal` or `/slice full` (this run only, never saved), then `plan_depth` from `acta config show`, then `full`.

Read only the file for the depth you picked, and never the other one:

| Depth | File | Read it when |
|---|---|---|
| `full` | [full.md](full.md) | the depth is `full` |
| `minimal` | [minimal.md](minimal.md) | the depth is `minimal` |

Depth changes only what a plan holds, never the gate. A minimal plan is saved, committed and shown (its path and one line a task) like a full one, and waits for a yes. A split (split.md) never starts a build.

## Scope Check

With `plan_split` on in `acta config show`, or `/slice split` for this run only (never saved), read [split.md](split.md) and cut the spec into several plans.

If the spec covers several independent subsystems, suggest one plan per subsystem. Each plan produces working, testable software on its own.

## File Structure

Before tasks, map the files to create or change and what each one is for. The structure drives the task decomposition.

- Each file has one clear responsibility and a clear interface. Prefer small, focused files.
- Files that change together live together. Split by responsibility, not by layer.
- In an existing codebase, follow its patterns and do not restructure unasked. A file you modify that has grown unwieldy may get a split in the plan.

## Debt items

A plan that works a debt item sets `parent: debt/<stem>` in its header and
lists every DBT id it closes in `closes:`, for example
`closes: [DBT-0017.01, DBT-0017.02]`. A plan that also finishes a scratch item
or another spec lists those in the same `closes:` list, so nothing is missed. `acta:land` ticks the debt items from that list after the merge.

## Task Right-Sizing

A task is the smallest unit that carries its own test cycle and is worth a
fresh reviewer's gate. When drawing task boundaries: fold setup,
configuration, scaffolding, and documentation steps into the task whose
deliverable needs them; split only where a reviewer could meaningfully
reject one task while approving its neighbor. Each task ends with an
independently testable deliverable.

A plan never holds a step that can only happen after landing; that work goes
in the landing report as the next action.

A plan holds 99 tasks at most, so task ids stay two digits wide. Work bigger
than that is several plans.

## Verify lines are properties

Every task carries a `verify:` line. Write it as a claim that must hold on every path, never as the one case someone saw:

| Instance (breeds review rounds) | Property (closes them) |
|---|---|
| "Return null when the row cannot be read" | "No failure path ever reports a consenting user as not consenting. List every path that can give a false or missing answer, and what each returns." |
| "The card makes no present-tense claim" plus a line range | "No surface the user sees claims this happens today. List every surface checked." |
| "Guard the old heading string" | "The old copy cannot come back in any form: reworded, partial, or next to the new copy." |

Never give a line range as the target; a range is an instance. Name the rule and ask for the list of paths checked.

Each case a verify line lists is named once, in the task whose layer owns the behaviour. A task for a layer above says it is wiring only: one success path and one error passed up from below. It lists no lower-layer case again. A case that crosses a trust boundary names the adversarial inputs to try in its verify line.

Inputs that run the same code path are one case. Every task names its `**Seam:**`, the public interface its tests call. The yes to the plan is the yes to its seams.

## Test commands

Every run step in a task names the narrowest command that proves it: the package or file the task touched, for example `go test ./internal/tui/ -run TestX` or `pytest tests/test_x.py::test_y`. Never `./...` in a task, and never the whole suite. The full suite runs once, in `acta:land`. Why: agents in other worktrees share the same machine, and each full run slows every other one down.

The header names both commands on a `**Tests:**` line: the fast one (a short mode, or the touched packages) and the full one. When the repo has `scripts/test`, they are `scripts/test` and `scripts/test --full`.

## Waves

Group tasks into waves by file ownership. Two tasks in one wave never touch the same file. A task that needs another task's output goes in a later wave. Declare the waves in the plan so the executor can run each wave's tasks in parallel.

## Keep it small

Unless `acta config show` says `coding_guide: off`, every plan's Global Constraints carry this line: "Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility." Tasks come from the user's ask only: no task for a helper, probe or list the user did not ask for. Review findings never become tasks here, except the one fix task `acta:review` asks for.

## Hand-off

Save the plan, show it to the user, and wait for a yes, whatever its depth. Approval of the design did not approve the plan.

Then run it with `acta:build`. First run `acta config show`. When it prints `build_executor: <name>`, that executor is chosen: name it and do not ask. A `(default)` mark or a missing line means `inline` (you write the code yourself); do not ask. The user can still pick `subagent` or `dispatch` (an omp agent in its own herdr tab) for one run.

**Subagent models.** Name no model for any subagent; the harness default picks it.
