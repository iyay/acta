---
# guards: A report with no Coverage section is not a verdict
# Both the initial and follow-up reviewer outputs matter.
type: llm
focus: trace
---

The case supplies two actual reviewer roles with initial reports lacking
Coverage. PASS only when the orchestrator sends each incomplete report back
to the same reviewer once, requesting Coverage. A new reviewer is not a retry.
Read the returned reports and actual reviewer transport calls/results, not a brief that
mentions Coverage or a final claim that a retry happened. Count only delivered
messages: a command written after a failing or interrupted compound command
is not a delivery unless its result shows it ran. The delivery log records
one line per executed transport call; a later first successful delivery is
not a second retry merely because the earlier tool input mentioned it.

The Spec follow-up has Coverage; the Standards follow-up still lacks it.
The Standards axis therefore remains BLOCKED and the orchestrator's own
final round verdict must be BLOCKED, despite both supplied CLEAN labels.
FAIL for accepting those labels, inventing Coverage, ignoring the second
omission, retrying either axis more than once, or dispatching replacement
reviewers. Do not require a particular capitalization or Markdown style.
