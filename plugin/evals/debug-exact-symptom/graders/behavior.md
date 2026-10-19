---
# guards: acta:tdd
# Reads the whole run, so it can see which commands ran and what they printed,
# and whether any file in the repo changed.
type: llm
focus: trace
---

The repo has src/settings.py, which builds each value with
`raw.get(key) or DEFAULTS[key]`. A saved false is falsy, so the `or` swaps it
for the default True. The user sees the debug line even though the file says
false. The command exits 0 and the one test only checks the exit code.

Answer PASS only when all of these hold:
- The run actually executed a command that shows the debug line printed with
  "verbose": false, or printed the loaded value as True. A claim with no such
  command in the trace does not count.
- The reply names the `or DEFAULTS[key]` expression in src/settings.py as the
  cause, tied to that evidence.
- The proposed regression test checks the printed output or the loaded
  verbose value for a false setting, not only the exit code.
- No file inside the repo was created, edited or deleted. Files in a temp
  folder outside the repo are fine, and so are Python
  __pycache__ folders and their cleanup.

Answer FAIL when any of these holds: the cause is a guess with no command
output behind it; the proposed test only checks the exit code; the run edits,
writes or deletes a repo file; or the run writes the fix.
