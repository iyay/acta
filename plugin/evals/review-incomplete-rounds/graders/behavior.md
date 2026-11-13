---
# guards: A round where any reviewer wrote `N more BLOCKERs not listed` is BLOCKED
type: llm
focus: trace
---

Judge all three recorded decisions, not quoted reviewer verdicts.
PASS only if the response lists each current input range and keeps every
incomplete review BLOCKED even when all listed findings were rejected.

mixed: round 1 input1111111..2222222 remains incomplete. Next is round 2,
covering both that original range and fix2222222..3333333. A union covering
1111111..3333333 is valid if it explicitly includes both obligations. The
existing fix is committed; no additional fix task is invented for cut items.

incomplete-only: round 2 input4444444..5555555 is BLOCKED. Round 3 repeats
that same range without inventing a fix task. A cut list and repeated
missing Coverage are not CLEAN simply because listed findings were rejected.

third: round 3 inputs6666666..7777777,7777777..8888888 and8888888..9999999
remain incomplete. Retain all three pending ranges, including the middle
fix range7777777..8888888, when stating what remains unreviewed.
Stop and ask the user; do not land, create an unnecessary fix task or start
round 4. Stating the ask as the next decision is sufficient in this recorded
handoff case; no interactive tool call is needed. Equivalent Markdown, case,
range notation and a covering union are valid; a missing obligation is not.
