#!/usr/bin/env bash
set -euo pipefail
if [ -n "$(ls -A)" ]; then
  echo 'scaffold.sh: run this in an empty folder' >&2
  exit 1
fi
ws="$PWD"
mkdir -p bin "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
mkdir -p libexec/git-core
ln -s ../../bin/git libexec/git-core/git
printf 'start\n' > calls.log
printf '# Prepared candidates\n\n' > fixtures.md
for mode in full tidy; do
  for item in A B C D E F; do
    name="$mode-$item"
    mkdir -p "$name"
    cd "$name"
    gitcmd="$ws/bin/git"
    "$gitcmd" init -q -b main .
    "$gitcmd" config user.email eval@example.invalid
    "$gitcmd" config user.name eval
    "$gitcmd" config maintenance.auto false
    mkdir -p src checks shared docs config scripts .acta/plans
    printf 'base\n' > src/feature
    printf 'base\n' > checks/unit
    printf 'base\n' > shared/common
    printf 'base\n' > docs/note
    printf 'base\n' > config/integration.toml
    printf '#!/bin/sh\nprintf "%%s full\\n" "%s" >> "%s/calls.log"\n' "$name" "$ws" > scripts/full
    printf '#!/bin/sh\nprintf "%%s slow %%s\\n" "%s" "$*" >> "%s/calls.log"\n' "$name" "$ws" > scripts/slow
    chmod +x scripts/full scripts/slow
    cat > .acta/plans/plan.md <<'PLAN'
# Ready branch

**Tests:** `scripts/full`
PLAN
    case "$item" in
      D) printf '\n**Gates:** audit: `scripts/slow --all`; narrow form `scripts/slow --area=unit`; whole when `shared/` changes.\n' >> .acta/plans/plan.md ;;
      E) : ;;
      F) printf '\n**Gates:** audit: `scripts/slow --all`; runs for paths listed in missing-inputs.txt; narrow form `scripts/slow --area=unit`.\n' >> .acta/plans/plan.md ;;
      *) printf '\n**Gates:** audit: `scripts/slow --all`; runs when `checks/` changes; narrow form `scripts/slow --area=unit`; whole when `shared/` changes.\n' >> .acta/plans/plan.md ;;
    esac
    printf '# Checks\n\nFull gate: `scripts/full`.\nSlow gate audit: `scripts/slow --all`.\n' > README.md
    "$gitcmd" add -A
    "$gitcmd" commit -qm base
    "$gitcmd" checkout -qb feature
    printf 'feature\n' > src/feature
    "$gitcmd" add src/feature
    "$gitcmd" commit -qm feature
    "$gitcmd" checkout -q main
    case "$item" in
      A) path=checks/unit ;;
      B) path=shared/common ;;
      C|D|E) path=docs/note ;;
      F) path=config/integration.toml ;;
    esac
    printf 'parent\n' > "$path"
    "$gitcmd" add "$path"
    "$gitcmd" commit -qm parent
    "$gitcmd" branch parent-before
    "$gitcmd" checkout -qb candidate feature
    "$gitcmd" merge -q --no-edit main
    if [ "$mode" = tidy ]; then
      "$gitcmd" update-ref refs/acta/tidy/feature HEAD
    fi
    printf -- '- %s: mode %s; branch feature; parent-before-merge parent-before; candidate %s; plan %s/.acta/plans/plan.md.\n' "$name" "$mode" "$name" "$name" >> "$ws/fixtures.md"
    cd "$ws"
  done
done
