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
# The repo asks for the quick review depth. The case checks that the review
# skill still runs two reviewers when the diff touches an auth check.
printf 'review_depth: quick\n' > .acta.yaml
cat > src/auth.py <<'PY'
def can_edit(user, doc):
    return user.id == doc.owner_id or user.is_admin
PY
cat > .acta/plans/2026-10-12-auth-fix.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-12 10:00:00"
hash: evdepth
---
# Auth Fix Implementation Plan

**Goal:** Let a document that is shared with everyone be edited by any user.

**Spec:** none

**Tests:** none, this repo has no test suite and no formatter.

### Task 1: Open editing for shared documents

**Files:**
- Modify: `src/auth.py`

**verify:** can_edit allows the owner and an admin, and also any user when the document is shared with all.

- [x] Code: allow any user when the document is shared with all
- [x] Commit: `auth: allow edit on shared documents`
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
# The branch is the finished work: one commit that changes who may edit.
# The plan and the setting stay on main, so the diff is only the auth check.
git checkout -q -b auth-fix
cat > src/auth.py <<'PY'
def can_edit(user, doc):
    return user.id == doc.owner_id or user.is_admin or doc.shared_with_all
PY
git add src/auth.py
git commit -qm "auth: allow edit on shared documents"
