---
created: "2026-10-08 08:23:11"
id: SPC-0110
hash: fu47spf
started: "2026-10-08 08:26:49"
finished: "2026-10-08 08:32:59"
---
# acta scratch add folds into its own last commit

Status: Bounded, approved by the user in chat on 2026-10-08.

Why: shape, probe and frame tell the agent to run `acta scratch add` after every answer, so a dead session loses nothing. Each call commits on its own (`internal/write/scratch.go:128` calls `finish`, which calls `gitc.Commit`). One brainstorm on 2026-10-08 made 7 commits `chore(scratch): add to 2026-10-07-repo-override-invisible` in four minutes, each 4 to 9 lines on the same file. The user wants one commit per item, not one per answer. The skills keep writing after every answer; only the commit changes.

Design:
- `internal/gitc` gets `CommitOrFold(repo, path, msg string, wasDirty bool) Result`. It runs the same checks as `Commit` (git repo, `wasDirty`, `busy`). Then, when every fold rule below holds, it amends HEAD with only `path` and keeps the message: `git commit --amend --only -m msg -- path`. When any rule fails, it does exactly what `Commit` does today.
- Fold rules, all must hold:
  1. HEAD's subject equals `msg`.
  2. HEAD changes only `path`, and HEAD has exactly one parent.
  3. HEAD is on a branch (not detached), and no remote-tracking branch contains HEAD (`git branch -r --contains HEAD` is empty).
  4. No other local branch contains HEAD, and no other worktree's HEAD is HEAD or a child of it. A build worktree that started from HEAD must keep its base.
  5. HEAD's author name equals `gitc.UserName(repo)`.
- When the amend itself fails, `CommitOrFold` falls back to a plain commit, so the text is never left out of git.
- `Result` gets a `Folded bool`, so tests can tell a fold from a new commit. `scratch add` has no `--json`, so nothing prints it and the command's output stays as it is.
- Only `scratch add` uses `CommitOrFold`. `scratch new`, `acta set`, bugs, debt, specs and plans keep one commit per call.
- No skill text changes. No config switch: folding is always on when `auto_commit` is on.
- Tests run in temp repos: two adds to the same item make one commit with both texts; each rule 1 to 5 broken on its own gives a second commit; a `wasDirty` file still is not committed.
- The last task adds 1 to the patch version in the three plugin files.

Wiki: after landing, add one line to `.acta/wiki/commit-subject-form.md` saying `scratch add` amends its own last commit under the rules above.
