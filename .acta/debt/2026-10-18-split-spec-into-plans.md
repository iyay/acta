---
id: DBT-0127
hash: druz5ws
parent: plans/2026-10-18-split-spec-into-plans
---
# Review NOTEs: Split Spec Into Plans Plan

- [ ] (medium) internal/gitc/gitc.go IsLinkedWorktree uses rev-parse --path-format=absolute (git 2.31+); on older git it returns false, so a worktree tick stamps split spec dates again and the second tidy land clashes.
- [ ] (medium) internal/cli/ready.go: outside a git repo, or with git missing, ready says the parent is not a branch and exits 2, though help calls that a load error (exit 3).
