#!/usr/bin/env bash
set -euo pipefail
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
mkdir -p "$HOME/.acta" .acta/specs reports refs bin .claude
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
export PATH="$PWD/bin:$PATH"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
python3 - <<'FIXTURE'
import json
from pathlib import Path
files = json.loads(r'''{".acta/specs/relay-a.md": "---\nid: SPC-0001\nhash: relay01\n---\n# Archive A\n## Owner\nA1: The owner decision is deferred.\n## Retention\nA2: The retention decision is deferred.\n## Delivery\nA3: The delivery decision is deferred.\n## Cleanup\nA4: The cleanup decision is deferred.\n## Failure\nA5: The failure decision is deferred.\n## Verification\nA6: The verification decision is deferred.\n", "reports/relay-a.txt": "1. A1 Owner decision. Spec .acta/specs/relay-a.md:7: \"A1: The owner decision is deferred.\" Consequence: implementation must guess the owner rule. Smallest clarification: choose that rule before planning.\n\n2. A2 Retention decision. Spec .acta/specs/relay-a.md:9: \"A2: The retention decision is deferred.\" Consequence: implementation must guess the retention rule. Smallest clarification: choose that rule before planning.\n\n3. A3 Delivery decision. Spec .acta/specs/relay-a.md:11: \"A3: The delivery decision is deferred.\" Consequence: implementation must guess the delivery rule. Smallest clarification: choose that rule before planning.\n\n4. A4 Cleanup decision. Spec .acta/specs/relay-a.md:13: \"A4: The cleanup decision is deferred.\" Consequence: implementation must guess the cleanup rule. Smallest clarification: choose that rule before planning.\n\n5. A5 Failure decision. Spec .acta/specs/relay-a.md:15: \"A5: The failure decision is deferred.\" Consequence: implementation must guess the failure rule. Smallest clarification: choose that rule before planning.\n\n6. A6 Verification decision. Spec .acta/specs/relay-a.md:17: \"A6: The verification decision is deferred.\" Consequence: implementation must guess the verification rule. Smallest clarification: choose that rule before planning.\n\n### Read\nSpec sections:\n- Owner\n- Retention\n- Delivery\n- Cleanup\n- Failure\n- Verification\nReference paths:\n- none\n", ".acta/specs/relay-b.md": "---\nid: SPC-0002\nhash: relay02\n---\n# Archive B\n## Owner\nB1: The owner decision is deferred.\n## Retention\nB2: The retention decision is deferred.\n## Delivery\nB3: The delivery decision is deferred.\n## Cleanup\nB4: The cleanup decision is deferred.\n## Failure\nB5: The failure decision is deferred.\n", "reports/relay-b.txt": "1. B1 Owner decision. Spec .acta/specs/relay-b.md:7: \"B1: The owner decision is deferred.\" Consequence: implementation must guess the owner rule. Smallest clarification: choose that rule before planning.\n\n2. B2 Retention decision. Spec .acta/specs/relay-b.md:9: \"B2: The retention decision is deferred.\" Consequence: implementation must guess the retention rule. Smallest clarification: choose that rule before planning.\n\n3. B3 Delivery decision. Spec .acta/specs/relay-b.md:11: \"B3: The delivery decision is deferred.\" Consequence: implementation must guess the delivery rule. Smallest clarification: choose that rule before planning.\n\n4. B4 Cleanup decision. Spec .acta/specs/relay-b.md:13: \"B4: The cleanup decision is deferred.\" Consequence: implementation must guess the cleanup rule. Smallest clarification: choose that rule before planning.\n\n5. B5 Failure decision. Spec .acta/specs/relay-b.md:15: \"B5: The failure decision is deferred.\" Consequence: implementation must guess the failure rule. Smallest clarification: choose that rule before planning.\n\n2 more findings not listed\n\n### Read\nSpec sections:\n- Owner\n- Retention\n- Delivery\n- Cleanup\n- Failure\nReference paths:\n- none\n"}''')
for name, text in files.items():
    Path(name).write_text(text)
FIXTURE
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval
git add .acta/specs reports refs
git commit -qm 'spec: record completed cold-read inputs'
