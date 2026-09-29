---
# guards: Spike and Bounded
type: llm
---

The change is one wrong word in one line, so it is Spike and Bounded work: a
short design in chat, a yes from the user, then the fix.

The Bounded path is not the Architectural path. A short spec in .acta/specs,
a plan, a worktree, a test, a review and a land all belong to the Bounded path
and are never a FAIL.

Answer PASS when the reply says the light path is enough for this change.

Answer FAIL only when the reply puts the typo through the Architectural path:
it files a scratch item to hang a brainstorm off, lays out competing
approaches with trade-offs, presents the design in sections and asks for
approval after each section, or says the change restructures how the parts fit
together or alters an interface.
