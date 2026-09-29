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

1. Preconditions, all of them: every task committed, nothing uncommitted in the worktree, the full test suite and type checks green with the output shown, and the review verdict CLEAN. Any one missing: report it, do not merge.
2. Plan progress: run `acta show <plan id>`. When `done` is less than `total`, do not merge: tick each box whose work is really done, and report the rest to the user.
3. Stray files: `git diff --name-only <base>..<head>` must list no `node_modules`, `.venv` or `.env*` path, and `git ls-tree -r <head>` must hold no symlink (mode `120000`) the plan did not add. Found: strip them from the branch first.
4. Parent: the branch recorded when the worktree was made. Unsure: ask the user; it is the one question allowed here. Before the merge, look at the main checkout: `git status --porcelain` must show no tracked change (untracked files do not count) and `git rev-parse --abbrev-ref HEAD` must print the parent. Either one fails: stop and report, and never check out, stash or reset to make it pass. Why: a main checkout sitting on another branch would take the merge, and the work on that branch would ride along with it.
5. From the main checkout: `git merge --no-ff <branch> -m "<what landed>. Verified at merge: <gate numbers>"`. Capture the main checkout path before leaving the worktree, since cleanup must run from outside it. A conflict: resolve it hunk by hunk by the intent of each side; never `--abort`, never drop a side. Cannot resolve it: stop, keep the worktree, report. Right after the merge, run `acta id --fix-duplicates` in the main checkout: two branches can take the same number, and this gives the later one the next free number (the hash never changes). Put each renumber line it prints in the landing report.
6. Run the gates again on the merge result. Red: say so plainly and leave the merge for the user.
7. Clean up: `git worktree remove <path>`, then `git branch -d <branch>`. `-d` refusing means not fully merged: stop and report. Never delete a worktree that has uncommitted work.
8. If the plan fixed a bug under `.acta/bugs/`, record the merge: `acta set bugs/<file-name-without-.md> fixed_in <merge-sha>`.
9. For each debt item in the plan's `closes:`, run `acta tick <DEBT-n.m> --all`. Then commit the ticks, because `acta tick` writes the debt file and does not commit it. A plan with no debt items in `closes:` skips this step.
10. Never `git push`, and never force anything.

Report after landing: the merge sha first, the gate numbers you ran after the merge, what was cleaned up, one next action, and, if step 9 ran, the debt file and how many items it closed.
