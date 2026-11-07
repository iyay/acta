---
id: DBT-0157
hash: hrazdfy
parent: plans/2026-11-07-stray-worktrees-rules
started: "2026-11-07 20:01:29"
---
# Review NOTEs: Stray worktrees: rules, land and tidy

- [x] (medium) tidy step 3 removes the temp gate worktree with a plain remove, so a gate run that leaves untracked files (node_modules, .venv, __pycache__) makes it refuse and stalls the tidy land on a user question.
