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
mkdir -p bin src .acta/plans
# The run will use git, and the system git does not work in the Bash sandbox. It
# is a stub that writes a cache file to a folder the sandbox locks. The real
# binary does not need that folder. A link would not do, because the sandbox
# cannot read the developer tools folder that the link points into. So the real
# git is copied.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
# acta starts git by name, and the system git may come first on PATH. So ./bin/acta
# is a small script that puts ./bin first and then starts the real binary. It holds
# the full folder name, because it cannot count on where it is run from.
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
# Task 1 of the plan is done, so this file is what that task left behind: it
# reads three columns and stops there.
cat > src/invoice.py <<'PY'
def invoice_total(line):
    row_id, item, qty = line.rstrip("\n").split(",")
    return {"id": row_id, "item": item, "qty": int(qty)}
PY
# A plan that work is going on right now: started, no finished stamp, task 1
# ticked, task 2 open, and a Next line that names the file and the change. The
# agent is given none of this; acta state prints it.
PLAN='---
depth: minimal
id: PLN-0042
created: "2026-10-05 09:00:00"
started: "2026-10-05 10:15:00"
hash: invtot
---
# Invoice Total Implementation Plan

**Goal:** invoice_total returns the line total as well as the three columns.

**Spec:** none

**Tests:** none, this repo has no test suite.

### Task 1: Read the three columns

**Files:**
- Create: `src/invoice.py`

**verify:** invoice_total returns a dict with the keys id, item and qty.

- [x] Code: read three columns and return the dict
- [x] Commit: `invoice: read the three columns`

### Task 2: Add the line total

**Files:**
- Modify: `src/invoice.py`
**verify:** invoice_total returns the key total, and it is qty times price.

- [ ] Code: read the fourth column, price, and put qty times price into the key total
- [ ] Commit: `invoice: add the line total`

## State

### Next

`src/invoice.py` reads three columns and stops. Read the fourth column, price,
and put qty times price into the key total in the dict it returns.

### Open rulings

none
'
printf '%s' "$PLAN" > .acta/plans/2026-10-05-invoice-total.md
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
# A plan only counts as running when it has a worktree of its own, so the plan
# and the work are on a linked checkout and this folder stays the main one.
git worktree add -q wt -b invoice-total
# The last thing the stopped session wrote down was this finding, so the branch's
# newest commit is its own work and not the scaffold.
printf '\n### Findings\n\nThe row is comma separated, so a price cannot hold a comma.\n' >> wt/.acta/plans/2026-10-05-invoice-total.md
git -C wt add .acta/plans/2026-10-05-invoice-total.md
git -C wt commit -qm "state: record what task 1 turned up"