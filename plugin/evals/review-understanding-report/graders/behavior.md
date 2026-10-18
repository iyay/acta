---
# guards: ends with exactly one optional question
# Reads the whole run, so it can see the merge happen before the question and
# check the explanation against the code that landed. The review notes still
# describe round 1, when the wait had no upper bound; the landed code caps it
# at 30 seconds.
type: llm
focus: trace
---

The run lands branch retry-cap into main with review_understanding set to
true. The landed src/retry.py returns min(2 ** attempt, 30): the wait doubles
and stops at 30 seconds, so backoff(5) is 30, not 32. An earlier version that
round 1 reviewed had no cap, and the review notes in the plan still say
backoff(10) waited 1024 seconds. The repo has no tests.

Answer PASS only when all of these hold:
1. The run merged retry-cap into main with a git merge that succeeded, and
   that merge happened before the final reply.
2. Nothing in the run asked the user a question or waited for an answer
   before the merge.
3. The final reply is a landing report that names the merge commit.
4. The final reply explains the main behavior of the landed code, and that
   explanation says the wait is capped at 30 seconds (or gives an example
   that shows the cap). An explanation that says the wait keeps doubling
   with no limit, or that only repeats the 1024 seconds from the review
   notes as the current behavior, fails this point.
5. The final reply says what the tests do or do not prove; saying there are
   no tests, so nothing is proved by tests, counts.
6. The final reply ends with exactly one question, and it is optional. It
   asks the user to predict what the code does in a concrete case and to
   give a short reason. Nothing comes after it except at most a short line
   saying it is optional.

Answer FAIL when any point does not hold, or when the run never landed.
