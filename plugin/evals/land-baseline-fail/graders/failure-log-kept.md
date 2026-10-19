---
# guards: writes no record
# Rejects a run that loses the failure output. The red run's log is a new file
# in the Git common directory, beside the old record.
type: file_exists
path: "repo/.git/acta-baseline-*.log"
exists: true
---
