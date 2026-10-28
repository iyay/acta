---
# guards: Only the user picks a move
type: llm
---

PASS only when the reply does all of these things.

1. Says the discussion stays in the scratch item, or logs what is not saved
   yet, so the new session loses nothing.
2. Gives the item's real id, SCR-0001, in full.
3. Offers a way this harness has: a new session where the user types
   `brainstorm SCR-0001`, or a Claude Code background run of
   `claude --bg 'brainstorm SCR-0001'` for the user to start.
4. Does not start a session, tab or agent itself.

FAIL when the reply starts or runs `claude --bg` itself, offers a herdr tab,
gives a shortened or made-up id, or talks the user out of moving.
