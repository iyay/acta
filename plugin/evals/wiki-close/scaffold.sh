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
mkdir -p bin src .acta/wiki .acta/plans
# The close runs git, and the system git does not work in the Bash sandbox. It is
# a stub that writes a cache file to a folder the sandbox locks. The real binary
# does not need that folder. A link would not do, because the sandbox cannot read
# the developer tools folder that the link points into. So the real git is copied.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
# acta starts git by name, and the system git may come first on PATH. So ./bin/acta
# is a small script that puts ./bin first and then starts the real binary. It holds
# the full folder name, because it cannot count on where it is run from.
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
cat > src/parse.py <<'PY'
def parse_row(line):
    row_id, name, qty = line.rstrip("\n").split(",")
    return {"id": row_id, "name": name, "qty": int(qty)}
PY
# The page is dated year 2000, so any commit on its file is newer than it.
cat > .acta/wiki/parse-output.md <<'PAGE'
---
type: Reference
title: parse_row output
description: parse_row returns a dict with the keys id, name and qty
paths: [src/parse.py]
timestamp: 2000-01-01T00:00:00Z
---
`parse_row(line)` reads one row, `id,name,qty`, and returns a dict with the
keys `id`, `name` and `qty`. `qty` is an int. The other two are text.
PAGE
# Every box is ticked, as it is when the last task of a plan is committed.
cat > .acta/plans/2026-10-04-parse-note.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-04 10:00:00"
hash: evclose
---
# Parse Note Implementation Plan

**Goal:** parse_row returns the new note column.

**Spec:** none

**Tests:** none, this repo has no test suite.

### Task 1: Add the note key

**Files:**
- Modify: `src/parse.py`

**verify:** parse_row returns a dict with a "note" key.

- [x] Code: read four columns and return the note key
- [x] Commit: `parse: add the note column`
PLAN
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
# The branch is the finished work. Its last commit changes src/parse.py, the
# file the page covers, and leaves the page as it was.
git checkout -q -b parse-note
cat > src/parse.py <<'PY'
def parse_row(line):
    row_id, name, qty, note = line.rstrip("\n").split(",")
    return {"id": row_id, "name": name, "qty": int(qty), "note": note}
PY
git add src/parse.py
git commit -qm "parse: add the note column"
