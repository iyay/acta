---
# guards: finish any outstanding type check
# Rejects a run that skips the one check the plan leaves open. The check
# leaves a marker file when it runs, so a reply cannot claim it ran.
type: file_exists
path: typecheck-ran
exists: true
---
