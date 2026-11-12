---
id: DBT-0164
hash: h0pxvq9
parent: plans/2026-11-12-land-gates-scoped
---
# Review NOTEs: Land runs the gates a diff needs, in any repo

- [ ] (medium) plugin/skills/land/baseline.md step 4 still says "An `eval:<case>` record follows the same rule" while land SKILL.md records `<gate>:<case>`; a repo whose gate has another name reads a wrong record form. The phrase is pinned in skill_land_test.go, so both change together.
- [ ] (medium) Land step 6 and tidy.md step 3 rerun a slow gate only when the parent gained a path the line says makes it run; the step 1 rules that a forcing path also makes it run and that a line with no run paths runs whole are not restated there, so a parent that gained only a forcing path can be read as "slow gate reused".
