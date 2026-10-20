---
# guards: No second fast suite
# Rejects a run that runs the repository-wide suite again at close. The suite
# leaves a marker file when it runs, so reading or quoting its command does
# not count as a run.
type: file_exists
path: broad-suite-ran
exists: false
---
