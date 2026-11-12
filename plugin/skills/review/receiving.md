## Receiving findings

Verify each finding against the code before acting. A wrong finding routed as a task wastes a round, and a right one you wave away ships a defect.

### Prove a BLOCKER first

Before a BLOCKER becomes a line of the fix task, run its repro yourself and keep the output.

1. Read the `file:line` it names, at the head of the range. A line the code does not hold settles it: `rejected: <reason>`.
2. Run the concrete input it gives. A command that writes, builds or reverts anything runs only in a temp clone (`git clone <worktree> <new temp folder>`), with `HOME` and `TMPDIR` set to new temp folders, never in the worktree.
3. The wrong result shows: it stays a BLOCKER, and the output goes with it into the fix task.
4. The wrong result does not show: its verdict is `rejected: <reason>`, with the output. It opens no fix task. Write it with the other findings of the round (`acta plan add <plan id> review-notes`).
5. No way to run it from here: say so, and ask the reviewer that raised it for a repro you can run. Until then it stays a BLOCKER; never drop it and never fix it blind.

A NOTE needs no repro, only a check that its `file:line` says what the NOTE claims.

### Every source passes the same check

| Finding from | Before you act |
|---|---|
| A reviewer subagent | Its report has a `### Coverage` section; then the steps above. |
| The fact-check | Each move names ground A or B and a line; read that line. |
| The user | Restate it, check it against the code, then act or push back. The user decides. |
| An outside reviewer | The steps above; it may not know the plan, so read the plan first. |
| Your own light inline review | The steps above; your own claim needs the same proof. |
| A `### Pre-existing` item | Read its `file:line` at the base of the range: the defect must be there before the diff. Then it goes to `acta:bug`. |

### Unclear feedback

If any item is unclear, stop: change nothing yet and ask about the unclear items. Items may hang together, so a part done on half an understanding is often wrong.

Example: the user says "Fix 1-6". You understand 1, 2, 3 and 6. Ask about 4 and 5 before you start on any of them.

### When to push back

Push back, with the code or the test output that shows it, when a finding:
- breaks something that works;
- comes from a reviewer that lacked the context;
- asks for a feature nothing uses;
- is wrong for this stack;
- undoes a choice kept for old data or old callers;
- goes against a design decision the user made. Bring that one to the user.

State the fact, show the output, ask a specific question. When the finding is right, fix it and say what changed and where.

### Order of work

One item at a time, each with its test, the BLOCKERs first. Cannot verify an item: say what you could not check and ask for direction; never go on as if it were checked.
