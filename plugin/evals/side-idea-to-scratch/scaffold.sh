#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
set -euo pipefail
# A first run with no voice file makes the session rules send the agent to
# /acta:setup before any other work, which spends a turn this case needs.
# A real machine is already set up; say so.
mkdir -p "$HOME/.acta"
printf 'chat_language: English\nstyle: adhd\nrepo_language: English\n' > "$HOME/.acta/voice.yaml"
mkdir -p bin .acta/scratch
cp "$(command -v acta)" bin/acta
git init -q .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
