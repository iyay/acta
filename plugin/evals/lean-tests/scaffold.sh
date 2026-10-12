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
mkdir -p bin src tests .acta/plans
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
# The two folders are empty packages, so the unittest command in the plan can
# import src.age and src.cli from the repo root.
: > src/__init__.py
: > tests/__init__.py
# Task 1 names its three cases. Task 2 is wiring only, so it names two paths.
# The case checks that the tests match this list and nothing more.
cat > .acta/plans/2026-10-12-age.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-12 10:00:00"
hash: leantx1
---
# Age Implementation Plan

**Goal:** A small command that reads an age from its arguments and prints it.

**Spec:** none

**Tests:** `python3 -m unittest discover -s tests -t .`

### Task 1: Parse an age

**Files:**
- Create: `src/age.py`
- Test: `tests/test_age.py`

**verify:** parse_age(text) returns the int for a whole number string, raises ValueError for a negative number, and raises ValueError for a non-number.

- [ ] **Failing test:** `tests/test_age.py` has three cases: parse_age("42") == 42; parse_age("-1") raises ValueError; parse_age("abc") raises ValueError.
- [ ] **Change:** `src/age.py` holds parse_age(text).
- [ ] **Commit:** `feat(age): parse an age`

### Task 2: Print the age

**Files:**
- Create: `src/cli.py`
- Test: `tests/test_cli.py`

**verify:** main(argv) prints the age parsed from argv[0]. This task is wiring only: one success path and one error passed up from below.

- [ ] **Failing test:** `tests/test_cli.py` has two cases: main(["42"]) prints 42; main(["abc"]) lets the ValueError from parse_age come up.
- [ ] **Change:** `src/cli.py` holds main(argv) and calls parse_age from `src/age.py`.
- [ ] **Commit:** `feat(cli): print the age`
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
