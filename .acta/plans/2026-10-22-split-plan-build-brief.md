---
parent: specs/2026-10-22-split-plan-build-brief-design
depth: minimal
id: PLN-0185
created: "2026-10-22 15:59:40"
hash: ndyj9nk
started: "2026-10-22 16:00:28"
finished: "2026-10-22 16:15:34"
---

# Split plans carry their own build brief

**Goal:** Each split plan holds a `## Build brief` with the non-default build context, and build reads it before it makes the worktree, so `build PLN-xxxx` alone is enough.
**Spec:** .acta/specs/2026-10-22-split-plan-build-brief-design.md
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Skill prose stays plain, short, and free of repo paths and user names (see `.acta/wiki/skill-text-fits-everyone.md`). Examples use made-up ids like `PLN-0160`.
- Every existing check in `TestPlanSplitText` and every skill size cap in `internal/plugincheck` keeps passing.

## Waves

- Wave 1: Task 1.

### Task 1: Write the brief in split and read it in build

**Files:** Modify `plugin/skills/slice/split.md`, `plugin/skills/build/SKILL.md`; test in `internal/plugincheck/skill_split_test.go`.
**Seam:** `TestPlanSplitText` over the skill text.
**verify:** split.md tells slice to add `## Build brief` to each split plan with only the four points that differ from the default (code repo, parent branch and whether it is checked out anywhere, land method, land order with siblings), to leave the section out when all four are default, to keep general skill rules out of it, and to keep the printed `build PLN-...` lines short. build/SKILL.md reads `## Build brief` before it picks the parent and makes the worktree, lets the repo and parent named there win over the defaults, and carries the land notes to `acta:land`. No other skill file changes. List each rule checked.
- [x] Red: `TestPlanSplitText` requires `## Build brief` in split.md and in build/SKILL.md; it fails because neither file has it.
- [x] Change: add the brief step to split.md before the print step, and one sentence in the `Parent:` paragraph of build/SKILL.md's Worktree section.
- [x] Commit: `feat(skills): split plans carry their own build brief`

## Polish

### Task 2: Review polish

**verify:** Each NOTE below applied, and nothing else changes.
- [x] split.md step 5 says to commit the brief with `acta commit` before the next step.
- [x] split.md step 5 gives land order its own "when" condition like the other three points.
- [x] Commit: `polish: review notes PLN-0185`

## Review notes

Deep review, full tier, two reviewers; round 1 CLEAN, polish (two [fix] NOTEs: commit the brief, land order condition) reviewed by two reviewers, CLEAN. One [debt]: a brief naming another code repo has no worktree, acta ready or land steps. Notes left: the test guards only the heading, byte caps sit at exact sizes, step 5 is dense.
