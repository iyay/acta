---
id: DBT-0142
hash: mqygu1g
parent: plans/2026-10-22-plan-completion-skill-gates
---
# Review NOTEs: Verify plan completion through build and land

- [ ] (medium) land-full-progress and land-tidy-progress graders pass any git merge-base --is-ancestor <sha> HEAD that exits 0, even with the tidy tip or HEAD itself, so an agent that skips acta commits still passes; require an acta commits call by plan hash before the check.
- [ ] (high) build-close-no-suite eval is red on both sides: on main 1/1 failed typecheck-ran, on this branch 3/3 failed no-fast-suite (the agent ran scripts/test at close). Suspect the build skill trim that dropped "a whole-repo run there spends every core on the machine"; rerun on main and on the branch before trusting the case.
