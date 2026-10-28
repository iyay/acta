---
id: DBT-0143
hash: utw8hjd
parent: plans/2026-10-28-tidy-branch-delete-no-confirm
---
# Review NOTEs: Exempt post-tidy branch delete from destructive rule

- [ ] (medium) TestLandTidyPath does not pin the step-7 no-confirmation sentence: reverting or rewording it stays green, so the land confirmation prompt can silently return. Pin "per core rule 5 this delete needs no confirmation" in internal/plugincheck/skill_land_test.go.
