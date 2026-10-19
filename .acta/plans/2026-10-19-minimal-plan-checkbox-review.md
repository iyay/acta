---
parent: bugs/2026-10-19-tick-all-fails-on-minimal-plan
depth: minimal
id: PLN-0177
created: "2026-10-19 20:27:17"
hash: gyars3k
---

# Check minimal task checkboxes before hand-off

**Goal:** Slice self-review catches missing Red, Change and Commit checkboxes in every minimal task before hand-off.
**Spec:** .acta/specs/2026-10-19-minimal-plan-checkbox-review-design.md
**Tests:** `scripts/test ./internal/plugincheck/`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Preserve spec coverage, property-shaped verify checks, approval gates and full-plan format; do not add a CLI validator or change tick behavior.
- Use an isolated worktree and the inline executor; keep skill text within existing size limits.

## Waves

- Wave 1: Task 1

### Task 1: Check required boxes during minimal-plan self-review

**Files:** Modify `plugin/skills/slice/SKILL.md` and `internal/plugincheck/skill_slice_test.go`.
**Seam:** `TestSkillSlice` checks the slice instructions through `CheckSkill`.
**verify:** Minimal self-review requires all three named checkbox steps for every task in every generated plan, corrects omissions before hand-off, and retains coverage and verify checks; the regression fails when that requirement is removed, and plugin size checks pass.
- [ ] Red: Extend `TestSkillSlice` to require the minimal self-review checkbox check and correction before hand-off; observe `scripts/test ./internal/plugincheck/ -run '^TestSkillSlice$'` fail against the current skill.
- [ ] Change: Extend the minimal self-review sentence with the required checkbox check for every task across all plans and correction before hand-off, then run the focused test and `scripts/test ./internal/plugincheck/`.
- [ ] Commit: `fix(slice): check minimal task boxes before hand-off`, with this task's Task trailer, after confirming that restoring the old skill makes the new regression fail.
