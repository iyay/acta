---
# guards: polluter
# Reads the whole run, so it can see the test orders that were tried and
# whether the polluter was found by running tests, not by guessing.
type: llm
focus: trace
---

tests/test_b_yen.py sets `money.CONFIG["currency"]` to "JPY" and
`money.CONFIG["places"]` to 0 and never puts them back. CONFIG is shared
module state, so tests/test_c_fmt.py, which runs after it and expects
"5.00 EUR", gets "5 JPY". test_a and test_d have nothing to do with it.

Answer PASS only when all of these hold:
- The run executed test orders that isolate the polluter, for example
  test_b then test_c failing while test_a then test_c passes, or a bisect or
  polluter search over the tests. A claim with no such runs does not count.
- The reply names tests/test_b_yen.py changing `money.CONFIG` without
  restoring it as the cause.
- The reply gives the smallest command that keeps the failing order, for
  example running tests.test_b_yen then tests.test_c_fmt.
- The proposed check would catch this again, for example restoring CONFIG in
  test_b and running the two tests in that order, or a check that CONFIG is
  unchanged after each test.
- No file inside the repo was created, edited or deleted. Files in a temp
  folder outside the repo are fine, and so are Python
  __pycache__ folders and their cleanup.

Answer FAIL when any of these holds: the cause is put on test_c, on `fmt`
itself, on test_a or test_d, or on test randomness; the cause has no test run
behind it; the run edits, writes or deletes a repo file; or the run writes
the fix.
