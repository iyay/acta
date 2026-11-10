#!/usr/bin/env bash
set -euo pipefail
# Fixtures must never replace a developer's files.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
mkdir -p "$HOME/.acta" .acta/specs cmd/greet
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cat > cmd/greet/main.go <<'GO'
package main

import (
	"flag"
	"fmt"
)

func main() {
	name := flag.String("name", "world", "who to greet")
	flag.Parse()
	fmt.Println("hello,", *name)
}
GO
# A small spec with no signal: one help string in one Go file, no auth, no
# data, no interface others use, one area.
cat > .acta/specs/greet-help-design.md <<'SPEC'
---
id: SPC-0001
hash: evnosig
---
# Clearer help for the name flag
Bounded.
## Problem
The help text of the `-name` flag in `cmd/greet/main.go` says "who to greet". It does not say the default.
## Design
Change the help text to "name to greet (default world)". Nothing else changes.
## Verification
`go run ./cmd/greet -h` shows the new help text.
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
git add .acta/specs/greet-help-design.md cmd/greet/main.go
git commit -qm "spec: greet help design"
