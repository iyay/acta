---
# guards: Related:
type: llm
---

PASS only when the reply does all of these things.

1. Keeps working on refunds in this session. Offering a new session is fine
   only as a choice with a concrete reason; this session may brainstorm again.
2. Treats refunds as its own topic with its own scratch item, or says it files
   one, instead of folding refunds into SCR-0001.
3. Names the earlier decisions that still apply to refunds: money in whole
   cents, and a sent invoice stays frozen.
4. Does not carry over a decision that does not touch refunds, for example it
   does not say refunds must use the billing table component.

FAIL when the reply says this session cannot start another brainstorm, sends
the user to another session as the only way, launches another session or agent
on its own, or merges refunds into the billing item.
