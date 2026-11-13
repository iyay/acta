---
# guards: temp worktree
# Actual file and root identities must remain unchanged.
type: regex
pattern: '^complete unchanged=1 root_unchanged=1\n$'
target:
  source: file
  path: audit.txt
---
