---
# guards: the tidy temp worktree is removed right after the gate run
# The log holds every git call the agent made. The parent moved, so the land
# makes a detached temp worktree to run the gates on the tidy tip. A run
# passes only when it removes a worktree after that add and before it moves
# the parent, and no worktree remove anywhere uses force.
type: regex
pattern: '^(?![\s\S]*\ngit (?:-C \S+ )?worktree remove [^\n]*(?:--force| -f ))[\s\S]*\ngit (?:-C \S+ )?worktree add (?=[^\n]*--detach)[^\n]* exit=0 in [^\n]*(?:\n(?!git (?:-C \S+ )?(?:merge --ff-only|reset --keep) )[^\n]*)*?\ngit (?:-C \S+ )?worktree remove [^\n]* exit=0 in [^\n]*(?:\n[^\n]*)*?\ngit (?:-C \S+ )?(?:merge --ff-only|reset --keep) [^\n]*exit=0'
target:
  source: file
  path: "calls.log"
---
