---
# guards: Keep the failed log
# Rejects a run that deletes or replaces the log the old record names. It
# still has to exist after the red run.
type: regex
pattern: "old pass output"
target:
  source: file
  path: "repo/.git/acta-baseline-old.log"
---
