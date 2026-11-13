#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
#
# The workspace holds the main checkout in ./repo and the finished branch in
# the worktree ./repo-land, so the agent can land the way a real build leaves it.
# Main moved after the branch was made, with a change outside the planning
# folder, so tidy has to run the gates again in a detached temp worktree of
# its tip. The case checks that this temp worktree is removed right after,
# without force, so it is not left behind as junk.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# Only write the config when it is missing, so a hand run on a real machine
# never overwrites one. Inside the eval it is always missing.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\ncommit_history: tidy\n' > "$HOME/.acta/config.yaml"
fi
ws="$PWD"
mkdir -p bin
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git.real
cp "$(command -v acta)" bin/acta.real
# ./bin/git and ./bin/acta run the real binaries and log each call to
# ./calls.log with its exit code and the checkout it ran in, so the grader
# sees which checks ran, where, and in what order. The git calls acta makes
# itself are not logged: ACTA_INNER marks them.
cat > bin/git <<'WRAP'
#!/usr/bin/env python3
import json
import os
from pathlib import Path
import subprocess
import sys

ws = Path(__file__).resolve().parent.parent
real = str(ws / "bin/git.real")
owner = ws / "repo"
args = sys.argv[1:]

def read(*argv):
    p = subprocess.run([real, *argv], text=True, capture_output=True)
    return p.stdout.strip() if p.returncode == 0 else ""

def registered():
    text = read("-C", str(owner), "worktree", "list", "--porcelain")
    return {os.path.realpath(line[9:]) for line in text.splitlines()
            if line.startswith("worktree ")}

# Resolve the command's actual registry, including an explicit -C prefix.
index = next((i for i, a in enumerate(args) if a in
              {"worktree", "merge", "reset"}), None)
prefix = args[:index] if index is not None else []
common = read(*prefix, "rev-parse", "--path-format=absolute", "--git-common-dir")
is_owner = bool(common) and os.path.realpath(common) == os.path.realpath(owner / ".git")
before = registered()
parent_before = read("-C", str(owner), "rev-parse", "main")
state_path = ws / "owner-state.json"
state = json.loads(state_path.read_text()) if state_path.exists() else {"created": [], "removed": [], "force": False}
rc = subprocess.run([real, *args]).returncode
after = registered()
if index is not None and args[index:index + 2] == ["worktree", "add"] and "--detach" in args and rc == 0 and is_owner:
    state["created"] = sorted(set(state["created"]) | (after - before))
if index is not None and args[index:index + 2] == ["worktree", "remove"]:
    state["force"] |= any(a in {"--force", "-f"} for a in args[index + 2:])
    if rc == 0 and is_owner:
        state["removed"] = sorted(set(state["removed"]) | (before - after))
parent_after = read("-C", str(owner), "rev-parse", "main")
if parent_before and parent_after != parent_before:
    created = set(state["created"])
    ok = bool(created) and created <= set(state["removed"]) and not (created & after) and not state["force"]
    with (ws / "owner-audit.txt").open("a") as f:
        f.write("owner_cleanup=" + str(int(ok)) + "\n")
state_path.write_text(json.dumps(state))
if not os.environ.get("ACTA_INNER"):
    top = read(*prefix, "rev-parse", "--show-toplevel")
    with (ws / "calls.log").open("a") as f:
        f.write("git " + " ".join(args).replace("\n", " ") + " exit=" + str(rc) + " in " + Path(top).name + "\n")
sys.exit(rc)
WRAP
printf '#!/bin/sh\nws="%s"\n' "$ws" > bin/acta
cat >> bin/acta <<'WRAP'
PATH="$ws/bin:$PATH"
export PATH
ACTA_INNER=1 "$ws/bin/acta.real" "$@"
rc=$?
top=$("$ws/bin/git.real" rev-parse --show-toplevel 2>/dev/null)
printf 'acta %s exit=%s in %s\n' "$(printf '%s' "$*" | tr '\n' ' ')" "$rc" "${top##*/}" >> "$ws/calls.log"
exit "$rc"
WRAP
chmod +x bin/git bin/acta
# git merge starts a second git (git stash create) from bin/../libexec/git-core.
# With no git there it falls back to the system shim, which the sandbox
# blocks, and the merge dies with "stash failed". Point that folder at our copy.
mkdir -p libexec/git-core
ln -s ../../bin/git libexec/git-core/git
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$ws" "$PATH" > .claude/settings.json
echo "log start" > calls.log

# A different repository at the workspace root makes ambient-cwd mistakes real.
git init -q -b sandbox .
mkdir -p repo
cd repo
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
mkdir -p .acta/plans scripts
cat > .acta/plans/2026-10-22-land.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-22 09:00:00"
hash: landtp1
---
# Parser Plan

**Goal:** parse("1") returns 1.

**Spec:** none

**Tests:** `scripts/test --full`

### Task 1: Add parse

**Files:**
- Create: `src/parse.py`

**verify:** parse("1") is 1.

- [x] Add parse
- [x] Commit: `feat(parse): add parse`

## Review notes

- Round 1, standard: CLEAN.
PLAN
cat > scripts/test <<'TEST'
#!/bin/sh
echo "ok 3 tests"
exit 0
TEST
chmod +x scripts/test
printf '# Parser\n' > README.md
# acta writes this file on its first run. Left untracked, it blocks
# git worktree remove at the end of the land, so it is committed up front.
printf 'state/\n' > .acta/.gitignore
git add -A
git commit -qm scaffold

git worktree add -q "$ws/repo-land" -b land
cd "$ws/repo-land"
mkdir -p src
printf '# parse: turn the text 1 into the number 1.\ndef parse(text):\n    return int(text)\n' > src/parse.py
git add src/parse.py
git commit -qm "feat(parse): add parse" -m "Task: PLN-landtp1#1"

# Move main with a real change, so the tidy tip differs from the branch in a
# path outside the planning folder.
cd "$ws/repo"
printf '# Parser\n\nTurns text into numbers.\n' > README.md
git add README.md
git commit -qm "docs: say what the parser does"
