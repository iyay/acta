---
# guards: the tidy temp worktree is removed right after the gate run
# The Git wrapper observes the owner's real registrations before and after
# each call. Only removal of every created gate worktree before the parent
# moves earns this record; the final reply cannot stand in for effects.
type: regex
pattern: '^owner_cleanup=1\n$'
target:
  source: file
  path: "owner-audit.txt"
---
