---
id: DBT-0138
hash: qzuiiso
parent: plans/2026-10-22-split-plan-build-brief
---
# Review NOTEs: Split plans carry their own build brief

- [ ] (medium) A build brief can name a code repo other than the plan repo, but build still puts the worktree at ../<repo>-<slug> of the plan repo, runs acta ready there, and land has no step that reads the brief (for example a parent checked out nowhere hits land step 4).
