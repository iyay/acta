#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
#
# The work is already landed on main. The case starts at the user's answer to
# the question that ended the landing report.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# The standard level with the understanding switch on. Only write the file
# when it is missing, so a hand run on a real machine never overwrites one.
# Inside the eval it is always missing.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nrepo_language: English\nreview_depth: standard\nreview_understanding: true\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin src .acta/plans
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
cat > .acta/plans/2026-10-18-retry-cap.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-18 10:00:00"
hash: evretry
---
# Retry Backoff Plan

**Goal:** A retry waits longer after each failure.

**Spec:** none

**Tests:** none, this repo has no test suite, no type checks and no formatter.

### Task 1: Add backoff

**Files:**
- Create: `src/retry.py`

**verify:** backoff(1) is 2 and backoff(3) is 8.

- [x] Code: add backoff
- [x] Commit: `retry: add backoff`

## Review notes

- Round 1, standard, two reviewers: CLEAN.
PLAN
cat > src/retry.py <<'PY'
def backoff(attempt):
    """Seconds to wait before retry number attempt. The first retry is 1."""
    # A long outage would make the wait grow past any useful size, so stop at 30.
    return min(2 ** attempt, 30)
PY
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
