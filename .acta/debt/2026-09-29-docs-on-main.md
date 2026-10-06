---
id: DBT-0023
hash: lv8rthg
parent: plans/2026-09-29-docs-on-main
---
# Review NOTEs: Specs and Plans on Main Implementation Plan

- [ ] (low) brainstorm/SKILL.md:130 says "commit on main" but :254 says "the branch that is checked out (usually main)"; plan and build say only "main". Pick one wording.
- [x] dispatch/SKILL.md:239 still says commit the plan on the branch as the first commit; with the plan already on main this is a no-op and now contradicts the new rule. (stale)
- [x] dispatch/SKILL.md:235 "if the plan is not yet committed on the branch" is dead text now. (stale)
- [x] dispatch/SKILL.md:216 makes the worktree from production; in a repo whose parent is not main the worktree lacks the spec and plan written on main. (stale)
- [x] build/SKILL.md:34 still says every change including docs gets a worktree; line 36 adds the spec/plan exception without calling it one.
- [ ] (low) plan/SKILL.md:16 "makes the worktree, only after the plan is approved": drop the comma.
- [ ] (low) land/SKILL.md:61 "after step 2 has committed the plan file" reads as if step 2 always commits; say "after step 2".
- [ ] (low) land/SKILL.md:61 is one long list item; the precondition re-check sentence is clunky next to its neighbours.
- [ ] (low) land/SKILL.md:61-62 an agent that stops at step 1's first sentence halts on a plan dirty from Close ticks; safe but a fake stop.
- [ ] (low) land/SKILL.md:62 "Then, when..." right after the done<total branch can read as belonging only to that branch.
- [ ] (low) land/SKILL.md:62 the step 2 tick commit lands after step 1 gates ran, so "Verified at merge" numbers are for a HEAD one plan-only commit older.
- [ ] (low) skill_land_test.go has no check that the tick commit is conditional beyond the porcelain string.
