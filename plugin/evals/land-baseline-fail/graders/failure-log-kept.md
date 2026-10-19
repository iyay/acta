---
# guards: writes no record
# Rejects a run that never names the new failure log. The report has to carry
# the absolute path of the red run's log, which has the time in its name, so
# the old record's relative log does not count.
type: regex
pattern: '(?:^|[^\w./-])/[^\n ]*acta-baseline-[0-9]{8}T[0-9]{6}Z[^ ]*\.log'
target: last_message
---
