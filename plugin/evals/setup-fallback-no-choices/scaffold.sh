#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# No config file is written on purpose: this case is a first run, so setup
# has every setting to ask.
mkdir -p bin
cp "$(command -v acta)" bin/acta
git init -q .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
