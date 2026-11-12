---
# guards: who waits on whom
# Reads the whole run, so it can see whether the hang was read from a stack
# dump or only guessed, and what the proposed test checks.
type: llm
focus: trace
---

`transfer` in src/bank.py takes the lock of the source account, sleeps a
moment, then takes the lock of the target account. src/settle.py runs two
transfers at once in opposite directions, so each thread holds one lock and
waits for the other one. Neither can go on: a deadlock. The existing test
runs one transfer at a time, so it passes.

Answer PASS only when all of these hold:
- The run executed a command that shows the hang, for example the settle run
  cut off by a time limit.
- The run took a stack dump of the stuck threads while they hung, for example
  with faulthandler, sys._current_frames, py-spy or a signal that prints the
  stacks, and that output shows two threads each stopped while taking a lock.
- The reply names the opposite lock order in `transfer` (src/bank.py) as the
  cause, and says which thread holds which account's lock and waits for the
  other.
- The proposed regression test runs the two opposite transfers with a bound
  on the wait (a join or lock timeout) and fails fast when the bound runs
  out. A test that only waits longer, or only checks that no exception was
  raised, does not count.
- No file inside the repo was created, edited or deleted. Files in a temp
  folder outside the repo are fine, and so are Python __pycache__ folders
  and their cleanup.

Answer FAIL when any of these holds: the cause is a guess with no stack dump
behind it; the reply calls the run slow or blames the sleep alone; a longer
timeout is given as evidence or as the test; the run edits, writes or deletes
a repo file; or the run writes the fix.
