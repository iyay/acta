# Probe mode

Ask the whole frontier instead of one question per message. Works on
the Bounded and Architectural paths.

**Map the decisions as a tree.** Every decision branches into the
decisions that hang off it. The *frontier* is every decision whose
prerequisites are settled: what you can ask now without guessing at
answers you have not heard yet.

**One round asks the frontier, five questions at most.** Put the
questions that unblock the most first; the rest wait for the next
round. A question whose answer depends on something still open in this
round waits too. Then stop and wait for the answers.

Each answer reshapes the tree: settled decisions push the frontier
outward. Recompute it after every round.

**Format each question** as a title, the body, then your answer, with no
code fence around the round:

**Q1. Scope of the retry**
Store the attempt count in Postgres, not in memory, so a restart does
not lose it. Which fits the deployment here?

Recommended: Postgres. The box restarts on deploy and an in-memory
count would hand the user a fresh budget every time.

**Facts are your job, never the user's.** When a frontier question
needs a fact from the environment (files, tools, config), dispatch a
subagent to find it under shape's subagent model rule. Do not wait on
it: only the questions downstream of that fact wait. The decisions
belong to the user, so put every one of them to them.

**On the Architectural path**, append each round's answers with
`acta scratch add SCR-0001 --section log`.

**Done when the frontier is empty**: every branch visited, nothing
left silently assumed. Ask the user to confirm you share an
understanding. Then shape goes on to approaches, or to the design.
