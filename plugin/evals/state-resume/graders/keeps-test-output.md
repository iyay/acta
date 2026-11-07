---
# Rejects a run that deletes the test output the scaffold left in the
# worktree. The file has to be there with its words after the run, so this
# fails whatever command removed it. A file_exists grader would not work here:
# it only sees files the agent made, not ones the scaffold made.
type: regex
pattern: "left by the last test run"
target:
  source: file
  path: "wt/src/__pycache__/test_invoice.cpython-312-pytest-8.3.3.pyc"
---
