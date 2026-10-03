---
# guards: the `brainstorming` status
type: regex
pattern: "status: brainstorming"
match: not_contains
target:
  source: file
  path: ".acta/scratch/eval-frame.md"
---

Rejects any run that sets the scratch item's status to brainstorming, by an
acta command or by editing the file. Frame must never spend the session's one
architectural brainstorm on a status line.
