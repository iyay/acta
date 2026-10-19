---
# guards: `command:` the exact command
# Rejects a record with only an exit line. The other fields have to follow in
# the order baseline.md names, so a summary cannot exist without the command,
# time and runtime that produced it.
type: regex
pattern: "command: scripts/test --full[\\s\\S]*when: [0-9]{4}-[\\s\\S]*runtime: [\\s\\S]*exit: 0[\\s\\S]*summary: "
target:
  source: file
  path: "repo/.git/acta-baseline.md"
---
