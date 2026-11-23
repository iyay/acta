---
# Cleanup never deletes or edits a test: removing one is a code change that
# goes through a plan. The text test is still there, word for word.
type: regex
pattern: "self\\.assertIn\\('print\\(\"usage: cli NAME\"\\)', source\\)"
target:
  source: file
  path: "tests/test_cli.py"
---
