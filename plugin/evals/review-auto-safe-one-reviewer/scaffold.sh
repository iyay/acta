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
# The branch adds one small function that nothing else uses yet, with a test,
# and the plan's task record shows that test failing before the function was
# written. No risk path is touched. The case checks that under the auto depth
# such a diff gets one reviewer subagent, not two and not none.
# The repo asks for the auto review depth, so the review picks the level.
printf 'review_depth: auto\n' > .acta.yaml
mkdir -p tests
cat > src/bar.py <<'PY'
def bar_width(width, percent):
    return width * percent // 100
PY
cat > .acta/plans/2026-11-12-clamp.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-11-12 10:00:00"
hash: evasafe
---
# Clamp a percent

**Goal:** A helper keeps a percent between 0 and 100.

**Spec:** none

**Tests:** `python3 -m unittest tests/test_clamp.py`; the repo has no formatter.

### Task 1: Clamp a percent

**Files:**
- Modify: `src/bar.py`
- Create: `tests/test_clamp.py`

**verify:** clamp_percent returns 0 for a value below 0, 100 for a value above 100, and the value itself in between. Nothing calls it yet.

- [x] Red: `tests/test_clamp.py` fails because `clamp_percent` is not defined (seen: NameError, 3 tests failed)
- [x] Change: add `clamp_percent` to `src/bar.py` (seen: 3 tests passed)
- [x] Commit: `bar: add clamp_percent`
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
# The branch is the finished work: the new function and its test.
git checkout -q -b clamp
cat > src/bar.py <<'PY'
def bar_width(width, percent):
    return width * percent // 100


def clamp_percent(value):
    if value < 0:
        return 0
    if value > 100:
        return 100
    return value
PY
cat > tests/test_clamp.py <<'PY'
import unittest

from src.bar import clamp_percent


class ClampPercent(unittest.TestCase):
    def test_below_zero(self):
        self.assertEqual(clamp_percent(-5), 0)

    def test_above_hundred(self):
        self.assertEqual(clamp_percent(140), 100)

    def test_in_range(self):
        self.assertEqual(clamp_percent(35), 35)


if __name__ == "__main__":
    unittest.main()
PY
git add src/bar.py tests/test_clamp.py
git commit -qm "bar: add clamp_percent"
