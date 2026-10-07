Read this when `commit_history` is `tidy`: it lands the branch in place of steps 5 to 8 of SKILL.md.

# Tidy landing

The gates of step 1 and the parent check of step 4 run as written. Then:

1. Pick the fold point (step 2) first. From the worktree: `acta tidy <parent> <branch>` (add `--onto <fold point>` when one folds). It prints `tidy: <old> commits -> <new>, tree ok, parent <sha>, folded <n>` and writes `refs/acta/tidy/<branch>`. A non-zero exit (a clash with the parent too) stops the land and reports the error; never fall back to a plain merge.
2. Fold point: `@{upstream}` of the parent, else `refs/acta/last-land`, else none. Only when every parent commit after it is a `chore(...)` commit does it fold. Any other commit there: no fold point. Pushed commits stay.
3. Gates: the tip is parent plus branch plus remapped hashes. When `git diff --name-only <branch> refs/acta/tidy/<branch>` lists only planning-root paths, write "only planning files differ, gates reused". Any other path (the parent moved): run the gates on the tidy tip first (a detached temp worktree of `refs/acta/tidy/<branch>`); red stops the land.
4. Read the new tip: `git rev-parse refs/acta/tidy/<branch>`. `fixed_in` takes it, never a hash from a planning file (tidy cannot rewrite a hash in its last commit).
5. Move the parent from the main checkout. First `git rev-parse --short=7 <parent>` must equal the `parent` sha of the `tidy:` line, else stop. `folded` is 0: `git merge --ff-only refs/acta/tidy/<branch>`. `folded` is above 0: `git reset --keep refs/acta/tidy/<branch>`; `--keep` refuses on a tracked local change: that stops the land. Then `acta id --fix-duplicates` and report each renumber line.
6. `git update-ref refs/acta/last-land HEAD` in the main checkout.
7. Clean up: `git worktree remove <path>`, then `git branch -D <branch>`: the old branch is no ancestor, and tidy proved the tip equals the merge of parent and branch. Then `git update-ref -d refs/acta/tidy/<branch>`.
8. Step 8 gets the commit from 4. Steps 9 and 10 as written.

In the landing report, the merge sha first is the new parent tip, and the `tidy:` line.
