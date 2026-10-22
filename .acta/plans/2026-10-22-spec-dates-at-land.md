---
parent: specs/2026-10-22-planning-writes-cheap-and-mergeable-design
depth: minimal
after: [PLN-0187]
id: PLN-0191
created: "2026-10-22 23:38:09"
hash: nqqbgb3
---

# Spec dates are written at land

**Goal:** A tick in a linked worktree never writes the spec, so merging a build cannot conflict with spec changes made on main, and land still gives every spec its dates.
**Spec:** .acta/specs/2026-10-22-planning-writes-cheap-and-mergeable-design.md
**Tests:** fast `scripts/test ./internal/write ./internal/cli ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Start the build only after PLN-0187 landed. It changes `plugin/skills/land/SKILL.md` and `internal/plugincheck/skill_land_test.go`.
- Plan dates do not move: tick still writes `started:` and `finished:` into the plan in any checkout. In the main checkout, tick still writes the spec dates.
- `SetField` keeps adding new keys at the end of the frontmatter. No merge driver.

## Waves

- Wave 1: Task 1, Task 2.

### Task 1: Tick leaves the spec alone in a worktree; spec-dates covers one-plan specs

**Files:** Modify `internal/write/mark.go` (`TaskDates`), `internal/write/specdates.go` (`SpecDates`); tests in `internal/write/specdates_test.go` and `internal/cli/specdates_test.go`.
**Seam:** `acta tick <plan>#task-N --start|--all` and `acta spec-dates <plan hash>`.
**verify:** No tick run in a linked worktree changes any spec file, for a spec with one plan or with several, on `--start`, `--step`, `--all` and `--undo`. List each path checked. In a git repo with a worktree, tick task 1 with `--start` and `--all` in the worktree and commit there, run `acta set <spec> status approved` on main, then `git merge --no-ff`: the merge is clean. Then `acta spec-dates <plan hash>` on main writes `started` from the merged plan and `finished` when every plan of the spec is done. It never moves an existing `started` later and never rewrites an existing `finished`. In the main checkout, tick still stamps the spec.
- [ ] Red: write the merge test above and a one-plan `SpecDates` case, run `scripts/test ./internal/cli -run TestSpecDates` and `scripts/test ./internal/write -run TestSpecDates`, and watch the merge conflict and the one-plan spec get no dates.
- [ ] Change: drop the `plansOf(...) > 1` condition from the linked-worktree check in `TaskDates`, drop the `len(plans) < 2` early return in `SpecDates`, rewrite both doc comments to say why, then rerun the two commands until they pass.
- [ ] Commit: `fix(write): write spec dates at land so worktree ticks never conflict`

### Task 2: Land text and wiki page match the new date rule

**Files:** Modify `plugin/skills/land/SKILL.md` (the `acta spec-dates` sentence in step 5), `.acta/wiki/tick-stamps-spec.md`, and the land check in `internal/plugincheck/skill_land_test.go` when it pins the old wording.
**Seam:** `scripts/test ./internal/plugincheck` and `acta wiki check`.
**verify:** No skill or wiki text says that tick in a worktree writes spec dates, or that `acta spec-dates` is only for a spec with several plans. Grep `plugin/skills` and `.acta/wiki` for `spec-dates`, `several plans` and `started` and list each hit checked. `scripts/test ./internal/plugincheck` passes and `acta wiki check` reports no error.
- [ ] Red: change the land check to require that `acta spec-dates` runs for every spec, run `scripts/test ./internal/plugincheck -run TestLand`, and watch it fail.
- [ ] Change: rewrite the land sentence and the wiki page, then rerun the plugincheck command and `acta wiki check` until both pass.
- [ ] Commit: `docs(land): spec dates come from land for every spec`

## Build brief

- Land order: either order. Sibling plan PLN-0190 lands on main too and shares no file with this plan.
