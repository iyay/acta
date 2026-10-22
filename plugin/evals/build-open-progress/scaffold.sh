#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
#
# Task 1 is done and committed. Task 2 is open and has no commit, but its
# heading slipped to "##", so acta counts one task and shows progress 1/1.
# A close that trusts done equals total starts the review on unfinished work,
# which is the bug this case guards.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no config file sends the agent to /acta:setup first. Only
# write it when it is missing, so a hand run never overwrites a real one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin src tests .acta/specs .acta/plans
# The system git may be a stub that the sandbox cannot run, so the real
# binary is copied. acta starts git by name, so ./bin/acta puts ./bin first
# on PATH and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
cat > .acta/specs/2026-10-22-shout-design.md <<'SPEC'
---
id: SPC-0001
created: "2026-10-22 09:00:00"
hash: shspec1
---
# Shout helpers

Status: Bounded, approved by the user in chat on 2026-10-22.

shout(text) returns the text in upper case with "!" at the end.
whisper(text) returns the text in lower case with "..." at the end.
SPEC
cat > .acta/plans/2026-10-22-shout.md <<'PLAN'
---
parent: specs/2026-10-22-shout-design
depth: minimal
id: PLN-0001
created: "2026-10-22 09:10:00"
hash: openpr1
---
# Shout Helpers Plan

**Goal:** shout and whisper helpers.

**Spec:** `.acta/specs/2026-10-22-shout-design.md`

**Tests:** fast `python3 -m unittest discover -s tests`

### Task 1: Add shout

**Files:** Create `src/shout.py`, `tests/test_shout.py`.

**verify:** shout("hi") returns "HI!".

- [x] **Failing test:** `tests/test_shout.py` asserts shout("hi") == "HI!".
- [x] **Change:** `src/shout.py` holds shout(text).
- [x] **Commit:** `feat(shout): add shout`

## Task 2: Add whisper

**Files:** Create `src/whisper.py`, `tests/test_whisper.py`.

**verify:** whisper("HI") returns "hi...".

- [ ] **Failing test:** `tests/test_whisper.py` asserts whisper("HI") == "hi...".
- [ ] **Change:** `src/whisper.py` holds whisper(text).
- [ ] **Commit:** `feat(shout): add whisper`
PLAN
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
# acta writes this file on its first run. It is committed up front so it never
# shows as stray work.
printf 'state/\n' > .acta/.gitignore
git add -A
git commit -qm scaffold
# Task 1 is built on the branch with its trailer. Task 2 has no commit.
git checkout -q -b shout
cat > tests/test_shout.py <<'TEST'
import unittest

from src.shout import shout


class ShoutTest(unittest.TestCase):
    def test_shout(self):
        self.assertEqual(shout("hi"), "HI!")
TEST
printf 'def shout(text):\n    return text.upper() + "!"\n' > src/shout.py
touch src/__init__.py
git add tests/test_shout.py src/shout.py src/__init__.py
git commit -qm "feat(shout): add shout" -m "Task: PLN-openpr1#1"
