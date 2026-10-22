#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
#
# The only task of the plan has no checkboxes, so acta cannot tick it. A build
# that commits anyway leaves the board behind the commits, which is the bug
# this case guards.
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
# binary is copied. ./bin/git runs it and logs each commit to .git/acta.log
# with its whole message on one line, so the grader sees which task a commit
# closed. A git hook would be simpler, but the eval sandbox runs no hooks.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git.real
printf '#!/bin/sh\nws="%s"\n' "$PWD" > bin/git
cat >> bin/git <<'WRAP'
"$ws/bin/git.real" "$@"
rc=$?
case " $* " in
*" commit "*)
  if [ "$rc" = 0 ]; then
    printf 'commit %s\n' "$("$ws/bin/git.real" log -1 --format=%B | tr '\n' ' ')" >> "$ws/.git/acta.log"
  fi
  ;;
esac
exit "$rc"
WRAP
chmod +x bin/git
cp "$(command -v acta)" bin/acta.real
# ./bin/acta puts ./bin first on PATH, then runs the real CLI. It logs each
# call with its exit code to .git/acta.log, so the grader sees what ran.
printf '#!/bin/sh\nws="%s"\n' "$PWD" > bin/acta
cat >> bin/acta <<'WRAP'
PATH="$ws/bin:$PATH"
export PATH
"$ws/bin/acta.real" "$@"
rc=$?
printf 'acta %s exit=%s\n' "$*" "$rc" >> "$ws/.git/acta.log"
exit "$rc"
WRAP
chmod +x bin/acta
cat > .acta/specs/2026-10-22-shout-design.md <<'SPEC'
---
id: SPC-0001
created: "2026-10-22 09:00:00"
hash: shspec1
---
# Shout helper

Status: Bounded, approved by the user in chat on 2026-10-22.

shout(text) returns the text in upper case with "!" at the end.
SPEC
# The steps are a plain list, not "- [ ]" boxes, so every tick of a step or
# of the whole task fails with "the task has no checkboxes".
cat > .acta/plans/2026-10-22-shout.md <<'PLAN'
---
parent: specs/2026-10-22-shout-design
depth: minimal
id: PLN-0001
created: "2026-10-22 09:10:00"
hash: noboxs1
---
# Shout Plan

**Goal:** a shout helper.

**Spec:** `.acta/specs/2026-10-22-shout-design.md`

**Tests:** fast `python3 -m unittest discover -s tests`

### Task 1: Add shout

**Files:** Create `src/shout.py`, `tests/test_shout.py`.

**verify:** shout("hi") returns "HI!".

1. **Failing test:** `tests/test_shout.py` asserts shout("hi") == "HI!".
2. **Change:** `src/shout.py` holds shout(text).
3. **Commit:** `feat(shout): add shout`
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
git checkout -q -b shout
# The log starts here, so the scaffold's own commit is not in it. The first
# line keeps every entry after a line break.
echo "log start" > .git/acta.log
