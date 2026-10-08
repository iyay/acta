#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no voice file makes the session rules send the agent to
# /acta:setup before any other work, which spends a turn this case needs.
# The eval home is a sealed throwaway, but a hand run in a normal shell would
# overwrite the voice file on a real machine, so only write it when it is
# missing. Inside the eval it is always missing, so the run still gets one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nrepo_language: English\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin
cp "$(command -v acta)" bin/acta
# The agent and acta commit both run git. The system git does not work in the
# Bash sandbox, because it is a stub that writes a cache file to a folder the
# sandbox locks. A link would not do, because the sandbox cannot read the
# developer tools folder that the link points into. So the real git is copied.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
git init -q .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
# The spec lands in its own commit, as acta id leaves it. The last commit then
# touches only the spec, so a second commit of the same file can fold into it.
mkdir -p .acta/specs
cat > .acta/specs/2026-10-08-x.md <<'SPEC'
---
created: "2026-10-08 09:00:00"
id: SPC-0001
hash: abc1234
---
# A small spec

The tool should run quickly on every file.
SPEC
git add .acta/specs/2026-10-08-x.md
git commit -qm "chore(spec): assign short ids"
