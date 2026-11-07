---
# guards: a dirty stray is never removed
# The last worktree list git printed must still hold repo-dirty.
type: regex
pattern: 'repo-dirty '
target:
  source: file
  path: "repo/.git/worktrees-now.txt"
---
