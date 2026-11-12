#!/usr/bin/env bash
set -euo pipefail
# Fixtures must never replace a developer's files.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# The spec says in one section that a finished archive is kept for 30 days and
# in another that it is deleted after 7 days. The case checks that the cold
# read finds that the two lines cannot both hold and names both.
mkdir -p "$HOME/.acta" .acta/specs docs
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cat > .acta/specs/retention-design.md <<'SPEC'
---
id: SPC-0001
hash: evcontr
---
# Export archive retention

## Storage
A finished export archive is stored in the workspace bucket.
A finished archive is kept for 30 days, so a user can download it again.

## Download
The owner gets a download link by email when the archive is ready.
The link works until the archive is gone.

## Cleanup
A nightly job deletes every finished archive 7 days after it was made.
The job logs each archive it deletes.
SPEC
mkdir -p bin .claude
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
export PATH="$PWD/bin:$PATH"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval
git add .acta/specs/retention-design.md
git commit -qm "spec: retention design"
