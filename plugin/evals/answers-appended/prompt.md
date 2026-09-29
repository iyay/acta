---
name: answers-appended
description: An answer the user gives lands in the scratch item body, not only in the chat.
tags: [brainstorm]
runs: 1
# The work is three steps: read the scratch skill, run the acta command, then
# answer the user. The command comes back with a git error in the sandbox, and
# the child may spend one or two more turns on it. Four kept runs of this case
# used 4, 5, 5 and 6 counted turns, so three is a coin flip. Seven is the most
# any run needed, plus one spare.
max_turns: 7
timeout_seconds: 240
allowed_tools: []
---

The idea "the download button fires twice when it is clicked fast" is already
filed in this repo as SCR-0001, in .acta/scratch/eval-answer.md.

My answer to the question in that item: it only happens in Safari, and only
when the click lands while the first request is still running.

Save that answer to the item with the acta command, then stop.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
