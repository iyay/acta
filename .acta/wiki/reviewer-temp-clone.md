---
type: Gotcha
title: Reviewers work in a temp clone, never a worktree
description: Reviewer worktrees leave files only --force removes and acta writes auto-commit, so reviewers use a temp clone
paths: [plugin/skills/review/]
---

Two things went wrong when reviewers worked inside git worktrees.

- A reviewer said it tested `acta debt new` in a temp copy, but it ran in the worktree and made two real commits.
- Reviewers ran `git worktree add /tmp/review-<sha>` and left untracked files in it (a test HOME with a read-only Go module cache, a goreleaser `dist/`). A plain `git worktree remove` refuses that, the pre-tool hook blocks `--force`, so every land left junk for the user.

Reviewer briefs say: a working copy of another revision, and any acta write command (debt new, bug new, set, id), go only in `git clone <repo> <tmp>`, from that folder. A clone is not in the repo's worktree list, so nothing is left for git to remove.

After each round the orchestrator runs `git worktree list` and reports any worktree the build did not make, and checks `git log <round head>..HEAD` for commits nobody dispatched. `acta worktrees` prints `run yourself: git worktree remove --force <path>` under a dirty stray.

`acta tick` writes the file but does not commit; commit ticks yourself.
