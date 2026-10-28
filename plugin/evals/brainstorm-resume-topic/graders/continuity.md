---
# guards: Resumed:
type: llm
---

PASS only when the reply does all of these things.

1. Shows it read SCR-0001: it restates what the item records, such as the
   SQLite full text search decision or twenty results at a time.
2. Asks the user to confirm at least one recorded assumption or decision
   before it relies on it, because things may have changed. Checking that the
   data still lives in one SQLite file counts.
3. Continues in this session. It does not say the topic must move elsewhere.

FAIL when the reply starts fresh as if nothing was recorded, builds on the
SQLite decisions without asking whether they still hold, says this session
cannot brainstorm, or launches another session or agent on its own.
