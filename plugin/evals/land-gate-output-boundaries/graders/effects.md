---
# guards: temp worktree
# The checker compares actual files to its original snapshots, not the reply.
type: regex
pattern: '^tracked unchanged=1 root_unchanged=1\npreexisting unchanged=1 root_unchanged=1\nunknown unchanged=1 root_unchanged=1\nsymlink unchanged=1 root_unchanged=1\nancestor-symlink unchanged=1 root_unchanged=1\noutside unchanged=1 root_unchanged=1\ntraversal unchanged=1 root_unchanged=1\nchanged unchanged=1 root_unchanged=1\nreplaced unchanged=1 root_unchanged=1\nworktree-identity unchanged=1 root_unchanged=1\nmissing-inventory unchanged=1 root_unchanged=1\nmixed unchanged=1 root_unchanged=1\nother-session unchanged=1 root_unchanged=1\nbuild unchanged=1 root_unchanged=1\nmain unchanged=1 root_unchanged=1\n$'
target:
  source: file
  path: audit.txt
---
