---
parent: bugs/2026-10-19-land-reads-code-before-merge
depth: minimal
id: PLN-0163
created: "2026-10-19 06:16:24"
hash: q99jk61
---
# Land read after merge plan

**Goal:** The understanding part of the landing report reads the landed code only after the merge succeeded.
**Spec:** `.acta/specs/2026-10-19-land-read-after-merge-design.md`
**Tests:** `scripts/test` (fast), `scripts/test --full` (full, land only); task loop: `scripts/test ./internal/plugincheck/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Change only the first bullet under `### Understanding, after the land`; no Go product code. Build inline in a worktree, English in the repo, never push.

## Waves

- Wave 1: Task 1.

### Task 1: Read the landed code after the merge

**Files:** Modify `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, and `internal/plugincheck/budget_test.go` only if the byte cap is exceeded.
**Seam:** `plugincheck.CheckSkill` with the land `SkillRule`; the eval case `review-understanding-report`.
**verify:** On every land path (plain merge and tidy) the explanation comes from a read made after the parent moved, from the parent checkout or `git show <landed-sha>:<path>`; a read of the old worktree's files or any read before the merge never counts. List each path and the line that owns it. `scripts/eval --case review-understanding-report --runs 3` passes 3 of 3.
- [ ] Red: Replace the old `Must` phrase with the new one and observe `scripts/test ./internal/plugincheck/ -run TestSkillLand` fail.
- [ ] Change: Rewrite the bullet; pass the same command and the eval run.
- [ ] Commit: `fix(land): read the landed code after the merge`.
