#!/usr/bin/env bash
# Makes a temp copy of the repo for debug probes and prints its path.
# Run it anywhere in the repo: it copies the whole work tree. The copy keeps
# uncommitted and untracked files, so the bug shows as it does in the repo.
# It leaves out the real .git and gives the copy an empty git of its own.
# On any failure it prints nothing and exits non-zero. Then
# `D=$(bash probe-copy.sh) && cd "$D"` stops before a probe writes anything.
set -euo pipefail

# Copy from the top of the git work tree, so a run from a subfolder still
# gets the whole tree. Outside git, copy the folder we are in.
from_git=no
if top=$(git rev-parse --show-toplevel 2>/dev/null); then
  src=$(cd "$top" && pwd -P) || exit 1
  from_git=yes
else
  src=$(pwd -P)
fi
base=$(cd "${TMPDIR:-/tmp}" 2>/dev/null && pwd -P) || exit 1

# A temp folder inside the repo would let probe writes land in the repo.
# Check before making anything, so even a failed run leaves the repo alone.
case "$base/" in
  "$src"/*) exit 1 ;;
esac

# Some sandboxes block mktemp but still allow a plain folder in TMPDIR.
dest=$(mktemp -d "$base/acta-probe.XXXXXX" 2>/dev/null) || {
  dest="$base/acta-probe-$$"
  # No -p: a folder or link someone left at this guessable name must make
  # us stop, not receive a copy of the repo.
  mkdir -m 700 "$dest" 2>/dev/null || exit 1
}

# Copy every entry, dotfiles too, except .git. In a linked worktree .git is
# a small file that points at the real repo, so git in a copy that kept it
# would move the real branch and index.
shopt -s dotglob nullglob
for entry in "$src"/*; do
  [ "${entry##*/}" = .git ] && continue
  cp -R "$entry" "$dest/" 2>/dev/null || exit 1
done

# Without a .git of its own, git in the copy would climb up the folders and
# could act on some other repo that holds the temp folder.
# A GIT_DIR set by a git hook would send git init to the real repo instead,
# so drop it for this one call.
if [ "$from_git" = yes ]; then
  (unset GIT_DIR GIT_WORK_TREE; git init -q "$dest") >/dev/null 2>&1 || exit 1
fi
printf '%s\n' "$dest"
