---
# The reply names the text test in tests/test_cli.py and the near duplicate
# pair in tests/test_price.py, each with a line number.
type: regex
pattern: '^(?=[\s\S]*tests/test_cli\.py:\d)(?=[\s\S]*tests/test_price\.py:\d)'
target: last_message
---
