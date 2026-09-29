#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
#
# The scratch item is seeded here under a fixed name, so a grader can read
# that one file and see whether the answer reached the body.
set -euo pipefail
# A first run with no voice file makes the session rules tell the agent to run
# /acta:setup before any other work, which spends a turn this case needs. A
# real machine is already set up; say so, or the case measures setup.
mkdir -p "$HOME/.acta"
printf 'chat_language: English\nstyle: adhd\nrepo_language: English\n' > "$HOME/.acta/voice.yaml"
mkdir -p bin .acta/scratch
cp "$(command -v acta)" bin/acta
cat > .acta/scratch/eval-answer.md <<'ITEM'
---
id: SCRATCH-1
hash: evalanswer
title: Download button fires twice
status: raw
created: 2026-01-01
---
The download button fires twice when it is clicked fast. Which browsers?
ITEM
git init -q .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
