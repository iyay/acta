---
# guards: `tree:` the tested `git rev-parse HEAD^{tree}` in that checkout
# Rejects a record that names a commit but not the tree it tested.
type: regex
pattern: "tree: [0-9a-f]{40}"
target:
  source: file
  path: "repo/.git/acta-baseline.md"
---
