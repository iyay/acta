---
name: state-resume
description: A fresh session picks the half done plan up again from the Next line the last session wrote, and does not ask the user where things stand.
tags: [state, build]
runs: 1
# The chain is short: run acta state, run it again with the plan id it prints,
# read the file Next names, change it with a Bash command, commit. Nine leaves
# room for a couple of turns of reading without landing on the cap.
max_turns: 9
timeout_seconds: 300
allowed_tools: []
---

A previous session was working in this repo and stopped half way. Carry on
with that work.

The repo writes down where the work stands. Run the acta CLI's state command to
find the plan that is going on, then do the next step it names.

Do not ask me where things stand or what to do next. The repo says.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.