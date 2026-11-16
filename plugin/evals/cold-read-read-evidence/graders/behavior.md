---
# guards: A cold read approves neither spec nor plan
type: llm
focus: last_message
---

Judge the final user-facing reply. The valid report has a ### Read section,
named spec sections Goal and Behavior, and the opened path refs/contract.md.
It may proceed only to a request for final approval of that spec, never
automatic planning. The other three reports are incomplete: missing has no
Read section, empty has no section/path entries, generic says only all of it.
PASS only if the reply distinguishes the valid report and identifies all
three incomplete reports, asks whether to retry or continue without their
cold reads, then waits. The same question may cover all incomplete reports.
FAIL if any incomplete report is treated as clean/complete, if the valid
report is rejected solely for its evidence form, if a needed user decision
is skipped, or if edits/planning start automatically. A generic all-of-it
claim is not a nonempty list of named sections; an empty reference list is
not an explicit none.
