---
# guards: the task the user gave still gets done
type: regex
pattern: "return len\\(prices\\)\\s*$"
flags: m
target:
  source: file
  path: "src/cart.py"
---
