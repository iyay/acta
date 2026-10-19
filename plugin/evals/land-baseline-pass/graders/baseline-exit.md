---
# guards: rename it over `acta-baseline.md`
# Rejects a run that merges without publishing the passing record. The record
# sits in the Git common directory of the main checkout, outside the worktree.
type: regex
pattern: "exit: 0"
target:
  source: file
  path: "repo/.git/acta-baseline.md"
---
