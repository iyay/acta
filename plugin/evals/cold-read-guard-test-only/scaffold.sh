#!/usr/bin/env bash
set -euo pipefail
# Fixtures must never replace a developer's files.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
mkdir -p "$HOME/.acta" .acta/specs skills/greet check
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cat > skills/greet/SKILL.md <<'SKILL'
---
name: greet
description: Use when the user asks for a greeting.
---
# Greet

Say hello to the user in a formal tone.
SKILL
# The only Go code here is a test that checks the skill text. It is not a
# second area, so the spec below must not get a cold read offer.
cat > go.mod <<'MOD'
module example.invalid/greet

go 1.22
MOD
cat > check/skill_test.go <<'GO'
package check

import (
	"os"
	"strings"
	"testing"
)

func TestGreetTone(t *testing.T) {
	b, err := os.ReadFile("../skills/greet/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "formal tone") {
		t.Error("greet skill must name its tone")
	}
}
GO
cat > .acta/specs/greet-tone-design.md <<'SPEC'
---
id: SPC-0001
hash: evguard
---
# Friendlier tone for the greet skill
Bounded.
## Problem
`skills/greet/SKILL.md` asks for a formal tone. Users find it stiff.
## Design
Change the sentence to "Say hello to the user in a friendly tone." Update the phrase that `check/skill_test.go` looks for from "formal tone" to "friendly tone". Nothing else changes.
## Verification
`go test ./check` passes with the new phrase.
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
git add .acta/specs/greet-tone-design.md skills/greet/SKILL.md check/skill_test.go go.mod
git commit -qm "spec: greet tone design"
