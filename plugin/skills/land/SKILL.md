---
name: land
description: "acta: Use when the review is CLEAN and every task is committed, and before claiming any work is done, fixed or passing. Runs the full gates with the output shown, checks the branch for stray files, merges --no-ff into the recorded parent without asking, removes the worktree and branch, and never pushes."
---

# Land

## Evidence before claims

### The Iron Law

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

If you have not run the verification command in this message, you cannot claim it passes.

### The Gate Function

```
BEFORE claiming any status or expressing satisfaction:

1. IDENTIFY: What command proves this claim?
2. RUN: Execute the FULL command (fresh, complete)
3. READ: Full output, check exit code, count failures
4. VERIFY: Does output confirm the claim?
  - If NO: State actual status with evidence
  - If YES: State claim WITH evidence
5. ONLY THEN: Make the claim

Skip any step = lying, not verifying
```

### Red Flags - STOP

- Using "should", "probably", "seems to"
- Expressing satisfaction before verification ("Great!", "Perfect!", "Done!", etc.)
- About to commit without verification
- Trusting agent success reports
- Relying on partial verification
- Thinking "just this once"
- Tired and wanting work over
- **ANY wording implying success without having run verification**

### Rationalization Prevention

| Excuse | Reality |
|--------|---------|
| "Should work now" | RUN the verification |
| "I'm confident" | Confidence is not evidence |
| "Just this once" | No exceptions |
| "Agent said success" | Verify independently |
| "I'm tired" | Tired is not an excuse |
| "Partial check is enough" | Partial proves nothing |
| "Different words so rule does not apply" | Spirit over letter |

## Landing

The approved plan already covers the merge. Do not show a menu (merge, PR, keep, discard) and do not ask.

1. Preconditions, all of them: every task committed, nothing uncommitted in the worktree, the full test suite and type checks green with the output shown, and the review verdict CLEAN. Run the full suite as `acta run-one -- <full command>` (`scripts/test --full` already does this), so only one full run uses the machine at a time. Any one missing: report it, do not merge. Read the clean-worktree precondition again after step 2 has committed the plan file, and remember that commit happens before the merge in step 5 and before `git worktree remove` in step 7. Eval gate: when the repo has scripts/eval and git diff --name-only <base>..<head> lists a path under plugin/skills/ or plugin/hooks/, run scripts/eval with the output shown; a red eval stops the land the same as a red test. Why: skill text is what agents obey, and only a real run shows they still do. Other repos and other diffs skip it. Wiki gate: from the worktree, before the merge, run `acta wiki check <base>..<head>` with the output shown; a problem stops the land the same as a red test. Fix the page in the worktree and commit it (`wiki.md` in `acta:build` has the rules), then check again. Most often a fix round touched the page's `paths`, so it only needs a `timestamp` bump. A repo with no wiki prints nothing.
State gate: before the merge, run `acta state <plan id>` and read its Findings; each one that still matters moves to a wiki page or to a debt item (`acta debt new`), so the next plan starts without it.
2. Plan progress: run `acta show <plan id>`. When `done` is less than `total`, do not merge: tick each box whose work is really done, and report the rest to the user. Then, when `git status --porcelain -- <plan path>` prints a line, commit the plan file in the worktree as `chore(plan): tick <plan>`, before the merge, so the ticks reach main. Build already commits ticks per wave, so often there is nothing to commit; then skip this.
3. Stray files: `git diff --name-only <base>..<head>` must list no `node_modules`, `.venv` or `.env*` path, and `git ls-tree -r <head>` must hold no symlink (mode `120000`) the plan did not add. Found: strip them from the branch first.
4. Parent: the branch recorded when the worktree was made. Unsure: ask the user; it is the one question allowed here. Before the merge, look at the main checkout: `git status --porcelain` must show no tracked change (untracked files do not count) and `git rev-parse --abbrev-ref HEAD` must print the parent. Either one fails: stop and report, and never check out, stash or reset to make it pass. Why: a main checkout sitting on another branch would take the merge, and the work on that branch would ride along with it.
   Mode: read `commit_history` from `acta config show` (a repo override shows there too). `tidy` (the default): follow "Tidy landing" below in place of steps 5 to 8. `full`: steps 5 to 8 as written.
5. From the main checkout: `git merge --no-ff <branch> -m "<what landed>. Verified at merge: <gate numbers>"`. Capture the main checkout path before leaving the worktree, since cleanup must run from outside it. A conflict: resolve it hunk by hunk by the intent of each side; never `--abort`, never drop a side. Cannot resolve it: stop, keep the worktree, report. Right after the merge, run `acta id --fix-duplicates` in the main checkout: two branches can take the same number, and this gives the later one the next free number (the hash never changes). Put each renumber line it prints in the landing report. `acta id` also rewrites ids still in the old format (`PLAN-30` becomes `PLN-0030`).
6. Run the gates again on the merge result, unless `git rev-parse HEAD^{tree}` prints the same tree as `git rev-parse <branch>^{tree}`: then the merge holds exactly the files the gates just passed, so write "tree same as branch, gates reused" in the report instead. Trees differ: list what changed with `git diff --name-only <branch> HEAD`; when every path sits under the planning root (`.acta/` is the default root; `.acta.yaml`, `ACTA_ROOT` or `acta --root` can move it, and the real root is the folder above `plans/` in the path `acta show <plan id> --path` prints), reuse the branch gates and write "only planning files differ, gates reused" in the report instead. Any other path, including Markdown outside the planning root, re-runs the gates. Red: say so plainly and leave the merge for the user.
7. Clean up: `git worktree remove <path>`, then `git branch -d <branch>`. `-d` refusing means not fully merged: stop and report. Never delete a worktree that has uncommitted work.
8. If the plan fixed a bug under `.acta/bugs/`, record the merge (in tidy mode the last task commit, see below): `acta set bugs/<file-name-without-.md> fixed_in <merge-sha>`.
9. For each debt item in the plan's `closes:`, run `acta tick <DBT-nnnn.nn> --all`. Then commit the ticks, because `acta tick` writes the debt file and does not commit it. A plan with no debt items in `closes:` skips this step.
10. Never `git push`, and never force anything.

### Tidy landing

With `commit_history: tidy`, the gates of step 1 and the parent check of step 4 run as written. Then:

1. From the worktree: `acta tidy <parent> <branch>`. It prints `tidy: <old> commits -> <new>, tree ok` and writes `refs/acta/tidy/<branch>`. A non-zero exit stops the land, leaves branch and worktree as they are, and reports the error. On any failure, never fall back to a plain merge.
2. Gates are not run again: the tree differs only by remapped hashes in planning files. Write "only planning files differ, gates reused" in the report.
3. Fold point: `@{upstream}` of the parent, else `refs/acta/last-land`, else none. When every parent commit after it is a `chore(...)` commit, run `acta tidy <parent> <branch> --onto <fold point>` in step 1 instead. They fold into the first commit that lands. Any other commit there: no fold point. Pushed commits are never touched.
4. Read the new tip now: `git rev-parse refs/acta/tidy/<branch>`. This is the last task commit, and `fixed_in` takes it. Tidy cannot rewrite a hash that sits in its own last commit, so never take it from a planning file.
5. Move the parent from the main checkout. Nothing folded: `git merge --ff-only refs/acta/tidy/<branch>`. Parent commits folded: check the parent tip is still the one tidy read, then `git reset --keep refs/acta/tidy/<branch>`. `--keep` refuses when a tracked file has local changes, and that stops the land. Then `acta id --fix-duplicates` and report each renumber line.
6. `git update-ref refs/acta/last-land HEAD` in the main checkout.
7. Clean up: `git worktree remove <path>`, then `git branch -D <branch>`: the old branch is no longer an ancestor, and tidy already proved the trees match. Then `git update-ref -d refs/acta/tidy/<branch>`.
8. Step 8 gets the commit from 4. Steps 9 and 10 as written.

Report after landing: the merge sha first (tidy mode: the new parent tip, and the `tidy:` line), the gate numbers you ran after the merge, what was cleaned up, the wiki pages the branch added or changed (`git diff --name-only <merge-sha>^1 <merge-sha> -- .acta/wiki/`, or none), one next action, and, if step 9 ran, the debt file and how many items it closed.
