---
# guards: temp worktree
# Actual file and root identities must remain unchanged.
type: regex
pattern: '^p1 unchanged=1 root_unchanged=1\np2 unchanged=1 root_unchanged=1\np3 unchanged=1 root_unchanged=1\np4 unchanged=1 root_unchanged=1\np5 unchanged=1 root_unchanged=1\np6 unchanged=1 root_unchanged=1\np7 unchanged=1 root_unchanged=1\np8 unchanged=1 root_unchanged=1\np9 unchanged=1 root_unchanged=1\np10 unchanged=1 root_unchanged=1\n$'
target:
  source: file
  path: audit.txt
---
