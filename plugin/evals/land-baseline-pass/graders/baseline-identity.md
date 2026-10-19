---
# guards: `commit:` the tested `git rev-parse HEAD`
# Rejects a record without the identity of what the suite tested. A bare exit
# line cannot say which tree passed.
type: regex
pattern: "commit: [0-9a-f]{40}"
target:
  source: file
  path: "repo/.git/acta-baseline.md"
---
