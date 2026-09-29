---
# guards: One Architectural brainstorm per session
type: llm
---

PASS only when the reply does all three of these things.

1. Says the second big change cannot start its brainstorm in this session,
   because this session already used its one Architectural brainstorm.
2. Offers the user one of two ways to run the second brainstorm elsewhere:
   a background agent, for example a `claude --bg` run of
   `brainstorm SCRATCH-n`, or a plain new session where the user types
   `brainstorm SCRATCH-n` themselves. Either one on its own is enough. The
   user only has to pick one, so do not ask for both and do not fail the reply
   because it offered one and left the other out.
3. Does NOT offer a herdr tab, does not name HERDR_ENV, and does not tell the
   user to open a new tab in some other tool.

FAIL when the reply starts the second brainstorm anyway, when it offers a
herdr tab as one of the choices, or when it leaves the user with no way to run
the second brainstorm elsewhere.

Do not fail the reply for anything this rubric never asked for. A wish of
your own that is not written above is not a FAIL.
