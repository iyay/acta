#!/usr/bin/env bash
set -euo pipefail
# Fixtures must never replace a developer's files.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# The spec says the limits file caps an export at 500 MB. The file it points at
# says 50 MB. The case checks that the cold read opens the reference and
# reports that the spec's claim about it is wrong.
mkdir -p "$HOME/.acta" .acta/specs docs
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cat > docs/limits.md <<'DOC'
# Limits

An export archive may be at most 50 MB.
A workspace may run one export at a time.
DOC
cat > .acta/specs/size-design.md <<'SPEC'
---
id: SPC-0001
hash: evclaim
---
# Export size check

## Problem
A large export fills the bucket and the user only learns of it at the end.

## Change
Before an export starts, add up the size of the records it will hold.
The limits file, [docs/limits.md](../../docs/limits.md), already caps an export archive at 500 MB.
Refuse an export above that cap with a message that names the cap.

## Testing
One test for an export under the cap and one for an export over it.
SPEC
mkdir -p bin .claude
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
export PATH="$PWD/bin:$PATH"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval
git add .acta/specs/size-design.md docs/limits.md
git commit -qm "spec: size design"
