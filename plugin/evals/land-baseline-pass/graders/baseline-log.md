---
# guards: `log:` the absolute path of the kept log
# Rejects a record that points at a relative path. A later worktree or a
# cleanup would make that path meaningless.
type: regex
pattern: "log: /"
target:
  source: file
  path: "repo/.git/acta-baseline.md"
---
