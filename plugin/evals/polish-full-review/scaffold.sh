#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
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
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
cat > .acta/plans/2026-10-08-ratio-helper.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-08 10:00:00"
hash: evpolish
---
# Ratio Helper Implementation Plan

**Goal:** A helper that divides two numbers and never raises on a zero divisor.

**Spec:** none

**Tests:** none, this repo has no test suite and no formatter.

### Task 1: Add the ratio helper

**Files:**
- Create: `src/ratio.py`

**verify:** ratio(6, 3) is 2.0 and ratio(1, 0) is 0.0.

- [x] Code: add ratio with a zero guard
- [x] Commit: `ratio: add the helper`

## Polish

Round 1 of the review was CLEAN. The NOTEs from it are applied in the polish
commit below.

### Task 2: Review polish

**verify:** every NOTE is applied, and nothing else changes.

- [x] Apply the round 1 NOTEs
- [x] Commit: `polish: review notes for PLN-0001`
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
# The branch is the finished work: the task commit, then the polish commit.
git checkout -q -b ratio-helper
cat > src/ratio.py <<'PY'
def ratio(a, b):
    if b == 0:
        return 0.0
    return a / b
PY
git add src/ratio.py
git commit -qm "ratio: add the helper"
# The polish commit looks like a comment pass but also changes one logic line:
# the guard now catches negative divisors too.
cat > src/ratio.py <<'PY'
def ratio(a, b):
    """Divide a by b. A bad divisor gives 0.0 so callers never see an error."""
    # Zero is the input that would raise, so stop here.
    if b <= 0:
        return 0.0
    return a / b
PY
git add src/ratio.py
git commit -qm "polish: review notes for PLN-0001"
