---
# guards: Read only until the cause is proven
# The fix waits for a plan, so the source file keeps its bug after the run.
type: regex
pattern: "return len\\(records\\) - 1"
target:
  source: file
  path: "src/count.py"
---
