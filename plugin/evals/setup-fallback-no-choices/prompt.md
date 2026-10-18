---
name: setup-fallback-no-choices
description: Setup in a session with no AskUserQuestion tool reads the config first, then falls back to the terminal wizard, and changes nothing on its own.
tags: [setup, claude-only]
runs: 1
# The eval sandbox never hands the child AskUserQuestion, even when
# allowed_tools names it, so this case can only check the fallback path.
# Load the skill, run doctor and config show, then reply.
max_turns: 6
timeout_seconds: 240
allowed_tools: [Skill, AskUserQuestion]
---

/acta:setup

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
