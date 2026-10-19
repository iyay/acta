---
# guards: no explanation, no offer
# The switch is false and the review ran deep. Deep must not turn the
# understanding part on, so the report stays the plain landing report.
type: llm
focus: last_message
---

This is the final reply after landing branch slug into main, with
review_understanding set to false and the review run at the deep level.

Answer PASS when all of these hold:
1. The reply names a merge commit, so the land happened.
2. The reply never shows what slug returns for some input, and never
   explains how slug treats spaces, case or other characters.
3. The reply never asks the user to predict or explain what the code does.
4. The reply never offers a quiz, check or question about how slug behaves.

Everything else in a landing report is fine and is not a reason to FAIL:
gate results (including that the repo has no tests, so no test ran),
review and plan checks, cleanup, sandbox or git warnings, open items about
files or ids, and a next action.

Answer FAIL when point 1, 2, 3 or 4 does not hold.
