---
parent: specs/2026-10-22-planning-writes-cheap-and-mergeable-design
depth: minimal
after: [PLN-0187]
id: PLN-0190
created: "2026-10-22 23:38:08"
hash: td8w75r
started: "2026-10-22 23:39:35"
---

# Plan writes through acta commands

**Goal:** Agents add a fix round, a polish task, review notes and an `after:` list with one acta command each, never with Read and Edit on the plan, and main cannot change a plan that is being built in a worktree.
**Spec:** .acta/specs/2026-10-22-planning-writes-cheap-and-mergeable-design.md
**Tests:** fast `scripts/test ./internal/write ./internal/cli ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Start the build only after PLN-0187 landed. It changes `plugin/skills/build/SKILL.md` and `internal/plugincheck/skill_build_test.go`.
- Every new write auto-commits through the same path as `acta state set` (`dirtyBefore` and `finish` in `internal/write/ops.go`), so the commit folds into the plan's last commit.
- Section names are exactly `fix-round`, `polish` and `review-notes`. The polish verify line is exactly `every NOTE below is applied, and nothing else changes.` and its commit box is exactly ``Commit: `polish: review notes for <plan id>` ``.
- Spec files are out of scope. No merge driver.

## Waves

- Wave 1: Task 1, Task 4.
- Wave 2: Task 2.
- Wave 3: Task 3.

### Task 1: `acta plan add`

**Files:** Create `internal/write/planadd.go`, `internal/write/planadd_test.go`, `internal/cli/plan.go`, `internal/cli/plan_test.go`; modify `internal/cli/cli.go` (dispatch case `plan`) and `internal/cli/help.go` (new entry in the Write group).
**Seam:** `acta plan add <plan> <section> [--title T] [--verify V] < stdin`.
**verify:** For each section, the file gains only the new lines, and every other byte stays the same. `fix-round` and `polish` tasks are numbered one past the highest `### Task N`, the round one past the highest `## Fix round`, and `acta tick` can tick every new box. New sections land before `## Review notes` and `## State` when present, otherwise at the end. `review-notes` adds lines and never replaces them. The command prints `plans/<stem>#task-N added`, or for `review-notes` the plan id. Inputs to try, each refused with no file change and no commit: unknown section, empty or blank-only stdin, `fix-round` without `--title` or `--verify`, a second `polish`, an id that is not a plan. Lines that start with `- [ ] ` or `- ` do not double the prefix. List each path checked.
- [x] Red: write the write-package and CLI tests for the cases above, run `scripts/test ./internal/write -run TestAddPlanSection` and `scripts/test ./internal/cli -run TestCmdPlanAdd`, and watch them fail.
- [x] Change: implement the writer, the command and its help entry, then rerun both commands until they pass.
- [x] Commit: `feat(cli): acta plan add writes fix rounds, polish and review notes`

### Task 2: `acta set <plan> after`

**Files:** Modify `internal/write/ops.go` (`SetValue`, case `after`), `internal/write/frontmatter.go` (write a flow list, remove on empty), `internal/cli/cli.go` (`cmdSet` validates with `badAfter` from `internal/cli/ready.go`), `internal/cli/help.go` (set usage lists `after`); tests in `internal/write/ops_test.go` and `internal/cli/cli_test.go`.
**Seam:** `acta set <plan> after <ids>`.
**verify:** A valid list, split on commas or spaces, is written as `after: [PLN-1, PLN-2]`. An empty value removes the field. Each `acta ready` exit-2 reason is refused with no file change: unknown id, not a plan, the plan itself, a dropped plan, a cycle. `after` on a spec, bug or scratch item is refused. List each path checked.
- [x] Red: write the tests, run `scripts/test ./internal/write -run TestSetValueAfter` and `scripts/test ./internal/cli -run TestCmdSetAfter`, and watch them fail.
- [x] Change: implement, then rerun both commands until they pass.
- [x] Commit: `feat(cli): acta set changes a plan's after list`

### Task 3: Guard a plan that is being built

**Files:** Create `internal/cli/guard.go` and `internal/cli/guard_test.go`; modify `internal/cli/cli.go` (`cmdSet`), `internal/cli/state.go` (`cmdStateSet`), `internal/cli/plan.go`.
**Seam:** `acta set`, `acta state set` and `acta plan add` run from the main checkout.
**verify:** No write command run outside a linked worktree changes a plan that `board.PlanStates` lists for a linked worktree. Each refusal prints `<plan> is building in <worktree>; change it there`, exits with the bad-input code and changes no file. The same command run inside that worktree still writes. A plan with no worktree, or one whose worktree is gone after land, still writes from main. List each command and path checked.
- [ ] Red: write the tests with a repo and a linked worktree that committed on the plan, run `scripts/test ./internal/cli -run TestGuard`, and watch them fail.
- [ ] Change: add one helper and call it from the three commands, then rerun until it passes.
- [ ] Commit: `feat(cli): refuse plan writes on main while a worktree builds it`

### Task 4: Skills use the commands

**Files:** Modify `plugin/skills/review/SKILL.md` (fix round step, polish step with the template block removed, review notes step), `plugin/skills/build/dispatch.md` (fix round step, run from the dispatched worktree), `plugin/skills/build/SKILL.md` (plan changes during a build only through `acta tick`, `acta state set` and `acta plan add`; no hand-written sections such as `## Verification evidence`), `plugin/skills/slice/split.md` (step 4 fixes the list with `acta set <plan> after`); checks in `internal/plugincheck/skill_review_test.go` and the build skill test.
**Seam:** `scripts/test ./internal/plugincheck`.
**verify:** No skill tells an agent to append to, or hand-edit, a plan section that `acta plan add` or `acta set` covers. Grep `plugin/skills` for `append`, `## Fix round`, `## Polish`, `## Review notes` and `after:` and list each hit checked. The review skill and `dispatch.md` name `acta plan add`, and plugincheck fails when either drops it. Skill size caps still pass.
- [x] Red: add the plugincheck checks, run `scripts/test ./internal/plugincheck -run 'TestReview|TestBuild|TestDispatch'`, and watch them fail.
- [x] Change: rewrite the skill lines, then rerun `scripts/test ./internal/plugincheck` until it passes.
- [x] Commit: `docs(skills): plan writes go through acta plan add and acta set`

## Build brief

- Land order: either order. Sibling plan PLN-0191 lands on main too and shares no file with this plan.

## State

### Next

Task 3: guard helper in internal/cli/guard.go, called from cmdSet, cmdStateSet and cmdPlan.
