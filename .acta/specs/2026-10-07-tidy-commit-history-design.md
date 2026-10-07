---
parent: scratch/2026-10-07-tidy-commit-history
id: SPC-0105
created: "2026-10-07 19:26:44"
hash: cmsh160
---
Status: Architectural, approved by the user in chat on 2026-10-07.
Why: acta auto-commits every planning change and lands with `merge --no-ff`, so a repo's history fills with `chore(...)` commits, review fix-ups and long `Land ...` merges. One code task should read as one commit. This repo went from 1984 commits to 452 with the same rules done by hand.

# Tidy commit history at land time

## Scope

In: a new user setting `commit_history`, a new command `acta tidy`, and a tidy path in the acta:land skill. Every repo that uses acta gets it, on by default.

Out: auto-commit stays as it is. Work in a worktree still commits every step, so nothing in flight is lost. Commits already pushed are never rewritten. Messages, trailers and authors are never changed.

## 1. Setting

- New key `commit_history: tidy | full` in the user config (`internal/config/user.go`), in the same shape as `plan_depth`, with a per-repo override. Default `tidy`.
- `full` keeps today's land: `git merge --no-ff` with the `Land ...` message.
- `acta config show` prints the key. `acta setup` asks for it.

## 2. `acta tidy <base> <branch> [--onto <fold point>]`

With `--onto`, first replays the `chore(...)` commits in `<fold point>..<base>` (see section 3) and folds them forward. Then walks the commits in `<base>..<branch>` in order (first parent on the branch side, merges skipped) and builds a new linear chain on top of `<base>`, or on top of the fold point with `--onto`.

- **Keep rule.** A commit stays a commit of its own when it touches a file outside the planning root and its subject type is not `chore`, `docs`, `polish` or `wiki`. Every other commit folds into the kept commit before it. When there is none yet, it rides along with the next kept commit.
- **Review fix-ups.** The first commit whose subject is a review-notes commit (`chore(plan): review notes ...`) or `polish` marks the end of the tasks. Every commit after it folds into the last kept commit.
- **Replay.** Each commit's own change is applied with `git merge-tree --write-tree --merge-base=<commit>^ <current> <commit>`. A clash folds that commit into the commit before it. The last commit always takes the exact tree of `<branch>`.
- **Dates.** Each new commit keeps the author date of its kept commit, or of its last folded commit when that is later. Dates never go backward along the chain.
- **Messages and authors.** Copied from the kept commit, unchanged.
- **Hash remap.** Old short or full hashes of `<base>..<branch>` commits found in planning-root files are replaced by the new commit that holds that change, at the same length. The change goes into the last new commit. Hashes that were never on `<branch>` stay as they are.
- **Proof.** The final tree must equal the tree of `<branch>` plus the remap edits, and nothing else. Any other difference: exit non-zero, write no ref, print why. `<branch>` itself is never moved.
- **Output.** The ref `refs/acta/tidy/<branch>` pointing at the new tip, and one line: `tidy: <old count> commits -> <new count>, tree ok`.

## 3. acta:land with `commit_history: tidy`

1. Gates run on the branch as today.
2. Run `acta tidy <parent> <branch>`. Non-zero stops the land and leaves the branch and worktree as they are. The tree only differs by remapped hashes in planning files, so the gates are reused and the report says so.
3. **Parent chore commits.** The fold point is `@{upstream}` of the parent, else `refs/acta/last-land`, else the parent tip (nothing folds). When every parent commit after the fold point is a `chore(...)` commit, `acta tidy --onto <fold point>` replays them first and folds them into the first commit that lands. Any non-chore commit there: tidy builds on the parent tip and folds nothing. Pushed commits are never touched.
4. Move the parent. Nothing folded: `git merge --ff-only refs/acta/tidy/<branch>`. Parent commits folded: from the main checkout, check the parent tip is still the one tidy read, then `git reset --keep refs/acta/tidy/<branch>`; `--keep` refuses when a tracked file has local changes, and that stops the land. Then `acta id --fix-duplicates` as today.
5. Set `refs/acta/last-land` to the new parent tip.
6. `fixed_in` gets the last new task commit, not a merge.
7. Cleanup removes the worktree and deletes the branch with `git branch -D`, since the old branch is no longer an ancestor. This is safe only because step 2 proved the trees match. Delete `refs/acta/tidy/<branch>`.

With `full`, land runs as today.

## 4. Docs and tests

- Wiki: rewrite `commit-subject-form` and `land-reuse-gates-planning-only` for the tidy path; new page `land-tidy-history` with the keep rule, the fold rules and why the branch is deleted with `-D`.
- Go tests on temp repos: keep rule, fold before and after the first kept commit, review fix-ups, a clashing replay, hash remap, a tree that differs stops tidy with no ref written, parent chore folding with and without an upstream, `refs/acta/last-land`.
- `internal/plugincheck`: the land skill names `acta tidy`.
- The last task bumps the plugin patch version in the three files.
