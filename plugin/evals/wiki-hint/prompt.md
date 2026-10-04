---
name: wiki-hint
description: A wiki page that covers a file reaches the agent when it touches that file, and the change follows what the page says.
tags: [wiki]
runs: 1
# The work is small: read the file, read the page the hint names, edit the
# file, answer. The session rules may send the child to a skill or two first,
# so ten leaves room for that without landing on the cap.
# The eval run grants Bash and no edit tool. A child that finds Edit off gives
# up and only shows the change, so the prompt says to write the file with Bash.
max_turns: 10
timeout_seconds: 300
allowed_tools: []
---

The design and the plan for this change are approved, and this folder is the
worktree, so the code goes here. One task is left.

src/parse.py has parse_row, which reads a row as `id,name,qty`. Rows now carry
a fourth column, `note`. Change parse_row so the dict it returns has a "note"
key as well.

The Edit and Write tools are off here, so change the file with a Bash command.
Then stop. No tests and no commit.
