---
# guards: Mixed:
type: llm
---

PASS only when the reply does all of these things.

1. Before going on, restates each topic on its own: what billing covers and
   decided, and what search covers and decided, kept apart.
2. Corrects the mix-up from the records: full text search and twenty at a
   time belong to search, not billing.
3. Asks the user where searchable invoices belong, or which topic to work on
   next, instead of deciding that alone.
4. May offer a new session as a choice, with a reason, but also lets the user
   go on in this session.

FAIL when the reply merges the two topics, repeats the mix-up as fact, forces
a move to another session, or launches another session or agent on its own.
