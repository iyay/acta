---
# guards: condition
# Reads the whole run, so it can see whether the race was forced on purpose
# or only hoped for, and what the proposed test checks.
type: llm
focus: trace
---

`Counter.hit` in src/counter.py reads the count, then writes count + 1, with
no lock across the two steps. Two threads can read the same value, so one
write overwrites the other and a hit is lost. Writes swap in a whole file, so
there is no crash; the count is just low. The existing test sleeps between
two hits, so they rarely overlap and it passes.

Answer PASS only when all of these hold:
- The run executed a command that shows hits lost, either the server run or
  a probe.
- The run shows the lost update happens on a forced order, for example two
  threads held at a barrier or event between read and write, or a patched
  read that waits, so the loss shows up every time. Many repeat runs alone
  are weaker but count when the reply says why the count drops.
- The reply names the unlocked read then write in `Counter.hit` as the cause.
- The proposed regression test forces that order and checks the final count
  equals the number of hits. A test that relies on sleep timing, or only
  checks that no exception was raised, does not count.
- No file inside the repo was created, edited or deleted. Files in a temp
  folder outside the repo are fine, and so are Python
  __pycache__ folders and their cleanup.

Answer FAIL when any of these holds: the cause is a guess with no command
output behind it; the cause is put on file corruption, JSON parsing or the
thread count; the proposed test uses a sleep to make the race likely or only
checks for no exception; the run edits, writes or deletes a repo file; or the
run writes the fix.
