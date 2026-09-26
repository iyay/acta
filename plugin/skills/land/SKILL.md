---
name: land
description: Use when the review is CLEAN and every task is committed, and before claiming any work is done, fixed or passing. Runs the full gates with the output shown, checks the branch for stray files, merges --no-ff into the recorded parent without asking, removes the worktree and branch, and never pushes.
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
2. Stray files: `git diff --name-only <base>..<head>` must list no `node_modules`, `.venv` or `.env*` path, and `git ls-tree -r <head>` must hold no symlink (mode `120000`) the plan did not add. Found: strip them from the branch first.
3. Parent: the branch recorded when the worktree was made. Unsure: ask the user; it is the one question allowed here.
4. From the main checkout: `git merge --no-ff <branch> -m "<what landed>. Verified at merge: <gate numbers>"`. Capture the main checkout path before leaving the worktree, since cleanup must run from outside it. A conflict: resolve it hunk by hunk by the intent of each side; never `--abort`, never drop a side. Cannot resolve it: stop, keep the worktree, report.
5. Run the gates again on the merge result. Red: say so plainly and leave the merge for the user.
6. Clean up: `git worktree remove <path>`, then `git branch -d <branch>`. `-d` refusing means not fully merged: stop and report. Never delete a worktree that has uncommitted work.
7. If the plan fixed a bug under `.pm/bugs/`, record the merge: `pmb set bugs/<file-name-without-.md> fixed_in <merge-sha>`.
8. Never `git push`, and never force anything.

Report after landing: the merge sha first, the gate numbers you ran after the merge, what was cleaned up, and one next action.
