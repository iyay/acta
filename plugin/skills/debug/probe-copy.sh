#!/usr/bin/env bash
# Makes a temp copy of the repo for debug probes and prints its path.
# Run it from the repo root. The copy keeps uncommitted and untracked files,
# so the bug shows as it does in the repo. It leaves out .git (see below).
# On any failure it prints nothing and exits non-zero. Then
# `D=$(bash probe-copy.sh) && cd "$D"` stops before a probe writes anything.
set -euo pipefail

src=$(pwd -P)
base=$(cd "${TMPDIR:-/tmp}" 2>/dev/null && pwd -P) || exit 1

# A temp folder inside the repo would let probe writes land in the repo.
# Check before making anything, so even a failed run leaves the repo alone.
case "$base/" in
  "$src"/*) exit 1 ;;
esac

# Some sandboxes block mktemp but still allow a plain folder in TMPDIR.
dest=$(mktemp -d "$base/acta-probe.XXXXXX" 2>/dev/null) || {
  dest="$base/acta-probe-$$"
  mkdir -p "$dest" 2>/dev/null || exit 1
}

# Copy every entry, dotfiles too, except .git. In a linked worktree .git is
# a small file that points at the real repo, so git in a copy that kept it
# would move the real branch and index.
shopt -s dotglob nullglob
for entry in "$src"/*; do
  [ "${entry##*/}" = .git ] && continue
  cp -R "$entry" "$dest/" 2>/dev/null || exit 1
done
printf '%s\n' "$dest"
