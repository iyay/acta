---
# The other half of the Next line: it names the fourth column, price, and the
# worktree's invoice.py never says the word, so a run that only read the state
# fails here too.
type: regex
pattern: "price"
target:
  source: file
  path: "wt/src/invoice.py"
---