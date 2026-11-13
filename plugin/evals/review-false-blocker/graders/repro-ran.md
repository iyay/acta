---
# guards: Run the concrete input it gives
type: regex
pattern: '^discount_result=40[.]0\n$'
target:
  source: file
  path: reports/repro.log
---
