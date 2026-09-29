---
name: scratch
description: "acta: Use when the user drops a raw idea (catet, nanti, kepikiran, note this, later) or a side idea shows up during other work. Files it in Scratchpad with acta scratch new; never in agent memory."
---

# Filing Ideas

A scratch item is a written idea in the repo, not a thing you keep in your head. File it in one command, tell the user in one line, then go back to what you were doing.

## When to file

- The user drops a raw idea: "catet", "nanti", "kepikiran", "note this", "later". File it at once. Do not ask first; the words are the idea.
- A side idea shows up while you build or fix something. That one is your guess, so ask first: "File this in Scratchpad?" Wait for the yes.
- Anything you must not forget, and the user did not ask for now.

## How to file

Write the user's words to a file, then run:

```bash
acta scratch new <slug> [--title T] < body.md
```

Then send one line and nothing else: `Filed SCR-0001 <title>`
To add to an item you already work on:

```bash
acta scratch add SCR-0001 [--section words|context|log|questions] < text.md
```

## What goes in

- The user's words, verbatim. Do not tidy or rewrite them. `new` puts them under `## Words`.
- Add context right after `new` with `acta scratch add SCRATCH-n --section context`: what work was going on, what already exists, the file:line spots, and where the facts came from (reading, debug, a run).
- Open questions go in with `--section questions`.
- More words from the user later go in with `--section words`, the default.
- Never write `---` or "Agent notes" by hand; the sections do that job.
- Pasted images are written as their paths, not as descriptions.
- Keep it short: an item is a hook for a later brainstorm, not a document.

## What never happens

- A scratch item never goes to agent memory. Files are the memory; the
  agent's notes are not.
- Filing an idea never needs a new session. The user keeps working and
  you keep working.
- A scratch write may commit on the main branch. These are data, not
  code, so they need no worktree, plan or review.

## Dropping an item

When the user says drop it, run `acta set scratch/<stem> status dropped`. Then say the item is dropped and move on. Dropped items stay in the repo so the history survives; they are just no longer active.
