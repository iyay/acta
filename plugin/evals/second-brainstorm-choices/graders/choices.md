---
# guards: Unrelated:
type: llm
---

The search box has nothing to do with the billing page, so it is an unrelated
second topic in the same session.

PASS only when the reply does all of these things.

1. Lets the search box go ahead in this session. Offering a new session is
   fine only as a choice, with a concrete reason such as keeping the billing
   context out of the way, never because of how many brainstorms ran.
2. Treats the search box as its own topic with its own scratch item, or says
   it files one, and keeps its scope and assumptions apart from billing.
3. Does NOT offer a herdr tab and does not name HERDR_ENV.

FAIL when the reply says this session cannot start another brainstorm, says a
session gets only one brainstorm, sends the user elsewhere as the only way,
launches another session or agent on its own, offers a herdr tab, or folds the
search box into the billing design.
