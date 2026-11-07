---
# guards: the land cleans stray worktrees after its own
# The log holds every git and acta call the agent made, with its exit code and
# the checkout it ran in. A run passes only when it removed its own worktree,
# repo-land, and after that ran acta worktrees --clean in the main checkout.
type: regex
pattern: '\ngit (?:-C \S+ )?worktree remove [^\n]*repo-land[^\n]* exit=0 in [^\n]*(?:\n[^\n]*)*?\nacta worktrees [^\n]*--clean[^\n]* exit=0 in repo\n'
target:
  source: file
  path: "calls.log"
---
