---
# Both tests of the near duplicate pair are still there.
type: regex
pattern: 'def test_total_with_tax\(self\):[\s\S]*def test_total_includes_tax\(self\):'
target:
  source: file
  path: "tests/test_price.py"
---
