---
name: frame
description: "acta: Use to frame a new product idea before designing it, when the user says frame, frame this idea, is this worth building, should I build this, new product idea, validate an idea or who is this for. Asks why, for whom and what, then writes the answers to a scratch item. Optional: a feature for something that already exists belongs to shape."
---

# Framing An Idea

Frame answers why, for whom and what. Shape answers how. Run frame only when
the user asks for it. It is optional and never forced; a feature for a product
that already exists stays with shape.

## 1. The scratch item

Use the scratch item the user names. Otherwise file their own words:

```bash
acta scratch new <slug> [--title T] < body.md
```

Never give the item the `brainstorming` status. The hook counts that status as
the session's one architectural brainstorm, and frame would spend it for
nothing. The status stays whatever `new` set, which is `raw`.

## 2. The goal question

Ask this first, before anything else. The answer decides how the rest of the
session runs.

> what's your goal with this?

Options: building a startup, or work inside a company; a hackathon or demo; open
source or research; learning; having fun.

Startup or work inside a company goes to startup mode: read `startup.md` and
use its six forcing questions. Everything else goes to builder mode, where you
improvise: ask what the most exciting version of the thing is, who they would
show it to, what would make that person say "whoa", what already exists that
comes closest, and what they would add with unlimited time. Builder mode is
collab, not interrogation; riff, then let the user cut it back.

When customers or revenue come up in builder mode, switch to startup mode
mid-session. Say so and read the file.

## 3. One question at a time

Each push-back builds on the last answer, so ask one question per turn and wait
for the reply. Never dump the whole list.

Skip any question the user's prompt already answered.

"Just do it", or a full plan with real evidence in it, means stop asking and go
to the premise challenge. The hard questions are the value; the plan is cheap.

## 4. The premise challenge

Before any direction, put the assumptions on the table as numbered premises the
user agrees or disagrees with, one at a time:

1. Is this the right problem? A different framing can give a much simpler
   answer.
2. What happens if nothing is built? Real pain or imagined pain.
3. What already exists? Look in the repo and name the patterns, helpers and
   flows that could be reused, and name the market tools that already do this.
4. For a new artifact, how do users get it? Code nobody can install is code
   nobody runs. Name the channel, or say it comes later on purpose.
5. In startup mode, does the evidence from the questions support this
   direction, and where is it thin?

If the user disagrees with a premise, change the understanding and put it back
to them. Do not argue past a yes.

## 5. Two or three directions

Give what, for whom and why. Never technical approaches; those are shape's job.

- The narrowest: the smallest thing someone would actually use.
- The ideal: where this goes if nothing limits it.
- One from a different angle: a different reading of the same problem, offered
  only when there is a real one.

Recommend one and say why in a line tied to the goal they gave you. The user
picks. Their pick is the answer, not a hint.

## 6. One assignment

End with a single concrete action in the real world: talk to three possible
users, watch one of them use the current workaround, find the ten people who
already do this by hand. Something that happens this week, outside this
session. Never "go build it".

## 7. Write it down

Write the result with `acta scratch add`, so the item is the record:

- `--section context`: who it is for, the problem, the agreed premises, the
  chosen direction.
- `--section log`: each answer, in the user's own words.
- `--section questions`: what is still open.

```bash
acta scratch add SCR-xxxx --section context < context.md
acta scratch add SCR-xxxx --section log < log.md
acta scratch add SCR-xxxx --section questions < questions.md
```

Nothing else writes to the item. No spec file, no plan.

## 8. Hand over to shape

Ask: "continue with shape SCR-xxxx now?" On yes, load `acta:shape` in this same
session and hand it the item. Shape already knows the why, so it starts on how.
On no, stop here; the item keeps the answers for later.

Frame never writes code, never picks a stack, and never turns into shape by
itself.