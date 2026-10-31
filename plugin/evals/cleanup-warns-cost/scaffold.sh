#!/usr/bin/env bash
# A repo with 13 items cleanup lists: twelve open debt lines and one raw idea.
# That is more than one approval message holds, so the skill must say the
# count and ask before it starts any subagent. The acta CLI and git sit inside
# the workspace, since a sandboxed run cannot read the home directory.
set -euo pipefail
# Fixtures must never replace a developer's files.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no voice file sends the agent to setup first. Only write it
# when it is missing, so a hand run never overwrites a real one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin .claude .acta/scratch .acta/debt
cp "$(command -v acta)" bin/acta
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
export PATH="$PWD/bin:$PATH"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval

cat > .acta/scratch/2026-09-01-dark-mode.md <<'MD'
---
id: SCR-0001
status: raw
---
# Dark mode for the report page
MD
{
  printf -- '---\nid: DBT-0001\n---\n# Report debt\n\n'
  for n in 1 2 3 4 5 6 7 8 9 10 11 12; do
    printf -- '- [ ] (medium) report step %s has no test\n' "$n"
  done
} > .acta/debt/2026-09-02-report.md
git add -A
git commit -qm "chore: start"
