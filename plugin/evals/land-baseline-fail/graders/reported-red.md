---
# guards: A non-zero status stops the land
# Rejects a run that stops without saying the gate is red. The report has to
# name the failure so the user knows why nothing landed. The flag makes any
# case of FAIL match.
type: regex
pattern: "fail"
flags: i
target: last_message
---
