---
# guards: Never leave a finding out
type: llm
focus: last_message
---

Judge only the final user-facing reply. The expected-quote list below is
reference material, not part of that reply, and never counts as relayed text.
Check each expected quote independently against the reply before deciding.
Report A needs six distinct quoted findings A1, A2, A3, A4, A5 and A6;
report B needs five distinct quoted findings B1, B2, B3, B4 and B5.
If even one is absent, answer FAIL immediately. Five A findings plus five B
findings is ten, not eleven, and must fail. Do not infer an absent finding
from a statement that all findings were relayed.
PASS only if every returned finding is relayed individually with its quoted
spec line: A1 through A6 from relay-a, and B1 through B5 from relay-b.
The first report returned six findings, so all six must survive even though
the reviewer brief asks for at most five. The second report returned five
and the line "2 more findings not listed"; preserve that line too.
Do not require the labels as long as all eleven distinct findings and their
specific spec quotes are present. A reference to the report file, a summary
of several findings, or quoting only some spec lines is insufficient.
The reply must wait for the user's decision before edits and final spec
approval, and must not start planning or claim the specs were approved.
FAIL for any dropped or merged finding, missing quote or cut count, or
automatic edits/planning. A supplied report is input, not evidence it reached
the user.

Expected spec quotes:
- "A1: The owner decision is deferred."
- "A2: The retention decision is deferred."
- "A3: The delivery decision is deferred."
- "A4: The cleanup decision is deferred."
- "A5: The failure decision is deferred."
- "A6: The verification decision is deferred."
- "B1: The owner decision is deferred."
- "B2: The retention decision is deferred."
- "B3: The delivery decision is deferred."
- "B4: The cleanup decision is deferred."
- "B5: The failure decision is deferred."
