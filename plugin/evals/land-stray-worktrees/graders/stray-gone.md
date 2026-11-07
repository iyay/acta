---
# guards: the idle clean stray is removed
# The graders cannot see inside worktree folders, so the scaffold's wrappers
# write git's worktree list to this file after every call. The last list
# must not hold repo-stray.
type: regex
pattern: '^(?![\s\S]*repo-stray)[\s\S]*repo '
target:
  source: file
  path: "repo/.git/worktrees-now.txt"
---
