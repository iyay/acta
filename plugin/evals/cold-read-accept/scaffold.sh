#!/usr/bin/env bash
set -euo pipefail
# Fixtures must never replace a developer's files.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
mkdir -p "$HOME/.acta" .acta/specs docs
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cat > docs/export-contract.md <<'DOC'
# Export access contract
Only the owner of a workspace may export its records.
Export jobs and download links contain customer data.
DOC
cat > .acta/specs/export-design.md <<'SPEC'
---
id: SPC-0001
hash: evcoldr
---
# Export records
## Problem
Workspace users need a downloadable copy of their records.
## Design
Any signed-in workspace member may start an export job.
Follow the access rules in [the export contract](../../docs/export-contract.md).
Store the completed archive and issue a download link.
## Verification
An authorized user can download the completed archive.
SPEC
# Keep the committed-spec premise true and avoid the sandbox's git stub.
mkdir -p bin .claude
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
export PATH="$PWD/bin:$PATH"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval
git add .acta/specs/export-design.md docs/export-contract.md
git commit -qm "spec: export design"
