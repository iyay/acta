---
type: Gotcha
title: Eval sandbox runs no git hooks
description: A post-commit hook in an eval scaffold never fires; log commits from a bin/git wrapper script instead
paths: [plugin/evals/]
---

A `.git/hooks/post-commit` hook that a scaffold installs never runs inside `claude plugin eval`. The same hook works when you run the scaffold by hand, so a grader that reads its log passes or fails for the wrong reason. Seen on 2026-10-23: an agent committed, and the grader saw no commit.

To log what an agent runs, copy git to `bin/git.real` and make `bin/git` a small script that runs it and appends to a log file. Do the same for `bin/acta`. See `plugin/evals/build-tick-failure/scaffold.sh` and `plugin/evals/land-full-progress/scaffold.sh`. Mark the git calls acta makes itself (an `ACTA_INNER` variable) when the grader must see only the agent's calls.

Two more traps from the same work:

- `scripts/eval --case` takes one case name. A glob like `build-[tm]*` finds nothing, and a second `--case` replaces the first. Give the cases a shared tag and run `--tag`.
- A `target: trace` regex sees each Bash command as a JSON string. Quotes inside the command are escaped, so `[^"]*` stops early. Match the body with `(?:[^"\\\n]|\\.)*`.
