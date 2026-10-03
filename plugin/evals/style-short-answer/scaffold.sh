#!/usr/bin/env bash
# The agent runs no tool in this case, so it needs no repo and no acta copy.
# It needs one thing: a config file in the home folder.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else, stop first, so a hand
# run never touches files that are not ours.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# With no config file, the session rules send the agent to /acta:setup before
# it answers, and the reply would be about setup. The eval home is a sealed
# throwaway, but a hand run in a normal shell would overwrite the config on a
# real machine, so only write it when it is missing. Inside the eval it is
# always missing, so the run still gets one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nrepo_language: English\n' > "$HOME/.acta/config.yaml"
fi
