---
# Passes when some python run failed while the test file was already changed
# and src/slug.py was not: the new row seen failing before the fix. The
# scaffold's python wrapper writes one line per run to this log.
type: regex
pattern: '^exit=[1-9][0-9]* src=orig tests=changed$'
flags: m
target:
  source: file
  path: ".git/test-runs.log"
---
