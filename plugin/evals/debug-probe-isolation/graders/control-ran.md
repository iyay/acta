---
# guards: Run a control
# Only a real run of the checker on a copy without the off-by-one prints this.
# The pattern reads the stdout of a shell call, so a reply that only says the
# words does not pass. The trace keeps stdout as a JSON string, so quotes
# inside it carry a backslash.
type: regex
pattern: 'stdout\\?":\s*\\?"(?:[^"\\\n]|\\.)*expected 3, got 3'
target: trace
---
