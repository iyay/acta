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
# A first run with no voice file makes the session rules send the agent to
# /acta:setup before any other work, which spends turns this case needs.
# The eval home is a sealed throwaway, but a hand run in a normal shell would
# overwrite the voice file on a real machine, so only write it when it is
# missing. Inside the eval it is always missing, so the run still gets one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nrepo_language: English\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin src .acta/wiki
cp "$(command -v acta)" bin/acta
# The parser splits a line on commas. The page below says that breaks on a
# quoted field. The prompt never says so, so the only way the agent learns it
# is the hint it gets when it touches this file.
cat > src/parse.py <<'PY'
def parse_row(line):
    row_id, name, qty = line.rstrip("\n").split(",")
    return {"id": row_id, "name": name, "qty": int(qty)}
PY
cat > .acta/wiki/parse-rows.md <<'PAGE'
---
type: Gotcha
title: Row fields can hold commas
description: A field can be quoted and hold commas, so read rows with csv.reader, never with line.split(",")
paths: [src/parse.py]
timestamp: 2026-10-04T00:00:00Z
---
The feed puts a field in double quotes when the field holds a comma, like
`7,"Lamp, small",3`. Splitting the line on commas cuts that name in two, and
the row then has too many fields.

Fix: read the line with `csv.reader`, which understands the quotes.
`next(csv.reader([line]))` gives the fields of one line.
PAGE
git init -q .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
