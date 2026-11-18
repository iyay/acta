---
# guards: core rule 11, the note stays as it was
type: regex
pattern: "^    # TODO\\(BUG-0007\\): free items still count here, see the bug\\.$"
flags: m
target:
  source: file
  path: "src/cart.py"
---
