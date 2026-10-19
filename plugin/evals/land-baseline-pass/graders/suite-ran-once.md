---
# guards: Never rerun the suite to recover a count
# Rejects a run that runs the suite twice. The scaffold counts runs in the Git
# common directory, and the second line says a summary cost another full run.
type: regex
pattern: "run 2"
match: not_contains
target:
  source: file
  path: "repo/.git/test-runs.log"
---
