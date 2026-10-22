---
# guards: errors block; stale warnings alone do not
type: regex
pattern: '^exit=0\n'
target:
  source: file
  path: "wiki-check-result.txt"
---
