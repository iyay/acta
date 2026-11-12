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
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin src .acta/plans
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
# The branch changes one line, a condition, and the repo has no test for it.
# No risk path is touched. The case checks that under the auto depth a small
# diff with an untested logic change still gets a reviewer subagent.
# The repo asks for the auto review depth, so the review picks the level.
printf 'review_depth: auto\n' > .acta.yaml
cat > src/discount.py <<'PY'
def apply_discount(price, percent):
    if percent < 0 or percent > 100:
        raise ValueError("percent must be between 0 and 99")
    return price * (100 - percent) / 100
PY
cat > .acta/plans/2026-11-12-discount.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-11-12 10:00:00"
hash: evarisk
---
# Reject a full discount

**Goal:** A discount of 100 percent is refused.

**Spec:** none

**Tests:** none, this repo has no test suite and no formatter.

### Task 1: Reject a full discount

**Files:**
- Modify: `src/discount.py`

**verify:** apply_discount raises ValueError for a percent of 100 or more, and for a percent below 0.

- [x] Code: refuse a percent of 100
- [x] Commit: `discount: refuse a full discount`
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
# The branch is the finished work: one condition changed, nothing else.
git checkout -q -b discount
cat > src/discount.py <<'PY'
def apply_discount(price, percent):
    if percent < 0 or percent >= 100:
        raise ValueError("percent must be between 0 and 99")
    return price * (100 - percent) / 100
PY
git add src/discount.py
git commit -qm "discount: refuse a full discount"
