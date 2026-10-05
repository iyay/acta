---
name: slice
description: "acta: Use after a design is approved, before any code. Turns the spec into a task-by-task plan in .acta/plans/ - failing test first, exact files, a property-shaped verify line and a wave for every task - then waits for the user's yes."
---

# Writing Plans

## Overview

Write the plan for a skilled implementer who knows nothing about this codebase, its tools or its domain, and who sees only their own task. Each task names the files to touch, the code, the tests and how to run them. Bite-sized tasks. DRY. YAGNI. TDD. One commit per task.

**Context:** Write and commit a new plan on main, next to its spec. `acta:build` makes the worktree, only after the plan is approved.

**Save plans to:** `.acta/plans/YYYY-MM-DD-<feature-name>.md` (`.acta/` is the default root; `.acta.yaml`, `ACTA_ROOT` or `acta --root` can move it).
- (User preferences for plan location override this default)
- Right after saving, run `acta id` right after so the plan gets its PLN number and hash before anyone refers to it.
- Commit the plan on main. The worktree branch starts from that commit, so it carries the plan.
- A fix round, or any change to a plan whose build is running, goes in that build's worktree, not on main.

## Plan depth

Pick the depth in this order, and do not ask when one answers: the argument of `/slice minimal` or `/slice full` (this run only, never saved), then `plan_depth` from `acta config show`, then `full`.

- `full`: everything below, with real code in every code step, and a wait for the user's yes.
- `minimal`: the frontmatter holds `parent:` and `depth: minimal`. The header is `**Goal:**` (one sentence), `**Spec:**`, `**Tests:**`, `## Global Constraints` (the lean line, unless coding_guide is off, plus only what this plan needs) and `## Waves`; no Architecture, Tech Stack, File map or Interfaces. Each task has a title, `**Files:**`, a property-shaped `**verify:**` and three one-sentence boxes: the failing test and why it fails, the code change, the commit message. No code blocks. About 6 lines a task.

The code-block rule and No Placeholders apply to `full` plans only. A minimal plan's self-review checks two things: every part of the spec has a task, and every verify line is a property.

A minimal plan needs no yes: save it, run `acta id`, commit it on main, tell the user its path in one line, and invoke `acta:build` in the same turn. The spec still needs its yes first.

## Scope Check

If the spec covers multiple independent subsystems, it should have been broken into sub-project specs during brainstorming. If it wasn't, suggest breaking this into separate plans — one per subsystem. Each plan should produce working, testable software on its own.

## File Structure

Before defining tasks, map out which files will be created or modified and what each one is responsible for. This is where decomposition decisions get locked in.

- Design units with clear boundaries and well-defined interfaces. Each file should have one clear responsibility.
- You reason best about code you can hold in context at once, and your edits are more reliable when files are focused. Prefer smaller, focused files over large ones that do too much.
- Files that change together should live together. Split by responsibility, not by technical layer.
- In existing codebases, follow established patterns. If the codebase uses large files, don't unilaterally restructure - but if a file you're modifying has grown unwieldy, including a split in the plan is reasonable.

This structure informs the task decomposition. Each task should produce self-contained changes that make sense independently.

## Debt items

A plan that works a debt item sets `parent: debt/<stem>` in its header and
lists every DBT id it closes in `closes:`, for example
`closes: [DBT-0017.01, DBT-0017.02]`. One piece of work can close more than what
it came from, so a plan that also finishes a scratch item or another spec
lists those in the same `closes:` list. One place, so nothing is missed:
`acta:land` ticks the debt items from that list after the merge.

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

## Test commands

Every run step in a task names the narrowest command that proves it: the package or file the task touched, for example `go test ./internal/tui/ -run TestX`, `pytest tests/test_x.py::test_y` or `npm test -- x.test.ts`. Never `./...` in a task, and never the whole suite. The full suite runs once, in `acta:land`. Why: agents in other worktrees share the same machine, and each full run slows every other one down.

The header names both commands on a `**Tests:**` line: the fast one (a short mode, or the touched packages) and the full one. When the repo has `scripts/test`, they are `scripts/test` and `scripts/test --full`.

## Waves

Group tasks into waves by file ownership. Two tasks in one wave never touch the same file. A task that needs another task's output goes in a later wave. Declare the waves in the plan, after the file map, so the executor can run each wave's tasks in parallel.

## Keep it small

Unless `acta config show` says `coding_guide: off`, every plan's Global Constraints carry this line: "Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility." Tasks come from the user's ask only: no task for a helper, probe or list the user did not ask for. Review findings never become tasks here, except the one fix task `acta:review` asks for.

## Bite-Sized Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement the minimal code to make the test pass" - step
- "Run the tests and make sure they pass" - step
- "Commit" - step

## Plan Document Header

**Every `full` plan starts with this header** (a `minimal` plan uses the shorter one in Plan depth):

```markdown
# [Feature Name] Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

**Spec:** [path to the spec/design doc this plan implements — the plan
argues from the spec, so the spec travels with it; executors read both.
A Bounded plan with no spec file writes exactly
`**Spec:** none (Bounded, approved in chat on <date>)`, with no other
backticks on the line: acta reads a .md path there as the spec]

**Tests:** [fast command, full command — see Test commands]

## Global Constraints

[The spec's project-wide requirements — version floors, dependency limits,
naming and copy rules, platform requirements — one line each, with exact
values copied verbatim from the spec. Every task's requirements implicitly
include this section.]

---
```

## Task Structure

````markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

**verify:** [the property that must hold on every path]

**Interfaces:**
- Consumes: [what this task uses from earlier tasks — exact signatures]
- Produces: [what later tasks rely on — exact function names, parameter
  and return types. A task's implementer sees only their own task; this
  block is how they learn the names and types neighboring tasks use.]

- [ ] **Step 1: Write the failing test**

```python
def test_specific_behavior():
    result = function(input)
    assert result == expected
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest tests/path/test.py::test_name -v`
Expected: FAIL with "function not defined"

- [ ] **Step 3: Write minimal implementation**

```python
def function(input):
    return expected
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest tests/path/test.py::test_name -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add tests/path/test.py src/path/file.py
git commit -m "feat: add specific feature"
```
````

## No Placeholders

Every step must contain the actual content an engineer needs. These are **plan failures** — never write them:
- "TBD", "TODO", "implement later", "fill in details"
- "Add appropriate error handling" / "add validation" / "handle edge cases"
- "Write tests for the above" (without actual test code)
- "Similar to Task N" (repeat the code — the engineer may be reading tasks out of order)
- Steps that describe what to do without showing how (code blocks required for code steps)
- References to types, functions, or methods not defined in any task

## Self-Review

After writing the complete plan, look at the spec with fresh eyes and check the plan against it. This is a checklist you run yourself — not a subagent dispatch.

**1. Spec coverage:** Skim each section/requirement in the spec. Can you point to a task that implements it? List any gaps.

**2. Placeholder scan:** Search your plan for red flags — any of the patterns from the "No Placeholders" section above. Fix them.

**3. Type consistency:** Do the types, method signatures, and property names you used in later tasks match what you defined in earlier tasks? A function called `clearLayers()` in Task 3 but `clearFullLayers()` in Task 7 is a bug.

If you find issues, fix them inline. No need to re-review — just fix and move on. If you find a spec requirement with no task, add the task.

## Hand-off

Save the plan, show it to the user, and wait for a yes (a minimal plan skips the wait, see Plan depth). Approval of the design did not approve the plan.

Then run it with `acta:build`. First run `acta config show`. When it prints `build_executor: <name>`, that executor is chosen: name it and do not ask. When the line is missing and the user has not said, ask which one: `subagent`, `dispatch` (an omp agent in its own herdr tab), or `inline` (you write the code yourself).

**Subagent models.** Only when `acta config show` lists `subagent_models: split` and you run in Claude Code: subagents that write code use `model: "sonnet"`; all other subagents (mapping, explore, planning help, debug investigation, spikes) use `model: "opus"`; reviewers use your own model alias. Otherwise name no model and follow the user's own config.
