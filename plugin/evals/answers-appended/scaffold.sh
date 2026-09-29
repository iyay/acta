#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
#
# The scratch item is seeded here under a fixed name, so a grader can read
# that one file and see whether the answer reached the body.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no voice file makes the session rules send the agent to
# /acta:setup before any other work, which spends a turn this case needs.
# The eval home is a sealed throwaway, but a hand run in a normal shell would
# overwrite the voice file on a real machine, so only write it when it is
# missing. Inside the eval it is always missing, so the run still gets one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nrepo_language: English\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin .acta/scratch
cp "$(command -v acta)" bin/acta
cat > .acta/scratch/eval-answer.md <<'ITEM'
---
id: SCR-0001
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
