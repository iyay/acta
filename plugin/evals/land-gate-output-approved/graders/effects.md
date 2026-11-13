---
# guards: temp worktree
# The checker compares actual files to its original snapshots, not the reply.
type: regex
pattern: '^approved removed=1 root_removed=1\n$'
target:
  source: file
  path: audit.txt
---
