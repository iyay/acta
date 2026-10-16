---
parent: specs/2026-10-18-reuse-full-suite-by-tree-design
depth: minimal
id: PLN-0157
created: "2026-10-18 19:12:11"
hash: dft1k8x
---
# Full Suite Only At Land Plan

**Goal:** The full suite runs only in `acta:land`; a review polish runs narrow tests and the formatter.

**Spec:** `.acta/specs/2026-10-18-reuse-full-suite-by-tree-design.md`

**Tests:** `scripts/test ./internal/plugincheck/`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Land step 6, the tidy gates, the eval gate and the wiki gate keep their current text.
- Skill prose stays general: no repo paths or `scripts/test` names in the review or land skill text.

## Waves

- Wave 1: Task 1, Task 2

### Task 1: Polish runs narrow checks

- [ ] **Files:** `plugin/skills/review/SKILL.md`, `internal/plugincheck/skill_review_test.go`
- [ ] **verify:** No polish path in the review skill asks for the full suite: the CLEAN-round step 2, the `--round polish` reply-back and every line that sends a polish to "step 3 onward" each name the narrow tests of the touched code, the formatter and `Review tiers`. The test fails if any of those lines says "full test suite", in any of the three places. List each line checked. Seam: `go test ./internal/plugincheck/ -run TestSkillReview`.
- [ ] Commit: `feat(review): polish runs narrow checks, land owns the full suite`

### Task 2: Land owns the full suite

- [ ] **Files:** `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `CLAUDE.md`
- [ ] **verify:** Land step 1 says the full suite runs after every change is in, that a red full run is fixed through the narrow failing tests and then run once more, and that the full suite is never a debugging loop; it runs again only after such a fix or when step 6 or tidy needs a new gate. `CLAUDE.md` Commands says the same: only as the land gate, never per task or per review round. The test requires the no-debugging-loop and run-again sentences. Seam: `go test ./internal/plugincheck/ -run TestSkillLand`.
- [ ] Commit: `feat(land): full suite runs once more only after a fix or a new gate`
