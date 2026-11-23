---
name: land
description: "acta: Use when the review is CLEAN and every task is committed, and before claiming any work is done, fixed or passing. Runs the full gates with the output shown, checks the branch for stray files, lands it on the recorded parent, no asking, removes the worktree and branch, never pushes."
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
- Expressing satisfaction before verification
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
| "I'm tired" | Tired is not an excuse |

## Landing

The approved plan already covers the merge. Do not show a menu (merge, PR, keep, discard) and do not ask.

Read [baseline.md](baseline.md) before every full suite run: step 1, a step-6 rerun on the merge result, or a tidy-tip rerun. It keeps the complete output and true exit status of every full run, and saves only a PASS as `acta-baseline.md` in the absolute Git common directory, shared by linked worktrees. Only a complete passing run replaces that record. A failed run, a reused gate or a write error keeps the last PASS and publishes nothing.

1. Preconditions, all of them: every task committed, nothing uncommitted in the worktree, the full test suite and type checks green with the output shown, and the review verdict CLEAN. In the plan worktree, untracked files a test or tool made (cache folders and the like) are not uncommitted work and are never deleted. Disposable temporary gate worktrees follow the approved-output cleanup procedure in tidy.md. Still there at land: show them and give the user three choices: add a `.gitignore` line (one commit on the branch; the gates then run again for that file, as step 6 does for a non-planning path), the user removes them, or stop the land. Never pick for the user. Run the full suite as `acta run-one -- <full command>` (`scripts/test --full` already does this), so only one full run uses the machine at a time. Any one missing: report it, do not merge. Run it once every change is in, polish included. Red: fix the cause through the narrow failing tests, then run it once more; never debug with the full suite. It reruns only then, or when step 6 or tidy needs a new gate. Exit 0 with `FLAKY:` lines: before the merge, record each one as [baseline.md](baseline.md) step 4 says. Read the clean-worktree precondition again after step 2 commits the plan file. Slow gates: the plan's `**Gates:**` line names each one, with its command, the paths that make it run, its narrow form and the paths that force the whole gate. For each gate, read `git diff --name-only <base>..<head>`: it runs only when the diff lists a path the line says makes it run, and a path that forces the whole gate also makes it run; it runs in the narrow form only when the line gives one and no listed path forces the whole gate; when it runs and the narrow form does not apply, or a narrow run picks no case, it runs whole, and so it does when you are in doubt about which paths the diff touches. A line that names no run paths means the gate runs whole. A plan with no `**Gates:**` line skips no gate: when the repo's own docs (`CLAUDE.md`, `AGENTS.md`, a README or a wiki page) name a slow gate, run it whole. Show the output of every gate run. Each red case runs again alone, picked the way the line or the gate's own tool names one case, up to 2 times, or once when an open bug lists that case under `flaky_tests` (files at the `path` values from `acta list --type bug --json`); never rerun the whole set for one case, it costs quota. A pass means the case is flaky: before the merge, run `acta flake record <gate>:<case> --plan <plan id>`, with the gate's name from the line, in the checkout [baseline.md](baseline.md) step 4 names, and go on. Red on every rerun stops the land. Wiki gate: from the worktree, before the merge, run `acta wiki check <base>..<head>` with the output shown; errors stop the land; stale warnings alone do not. Read affected pages and diffs; fix incorrect facts in the worktree and commit changes (`wiki.md` in `acta:build`), then check again. `acta wiki ok <id>` is optional after review. A repo with no wiki prints nothing.
State gate: before the merge, run `acta state <plan id>` and read its Findings; each one that still matters moves to a wiki page or to a debt item (`acta debt new`; each debt line starts with `(medium) ` or `(high) `), so the next plan starts clean.
2. Plan progress: run `acta show <plan id>`. When `done` is less than `total`, do not merge: tick each box whose work is really done, and report the rest to the user. Then, when `git status --porcelain -- <plan path>` prints a line, commit the plan file in the worktree with `acta commit <plan path> -m "chore(plan): tick <plan>"`, before the merge, so the ticks reach main.
3. Stray files: `git diff --name-only <base>..<head>` must list no `node_modules`, `.venv` or `.env*` path, and `git ls-tree -r <head>` must hold no symlink (mode `120000`) the plan did not add. Found: strip them from the branch first.
4. Parent: the branch recorded when the worktree was made. Unsure: ask the user; it is the one question allowed here. Before the merge, look at the main checkout: `git status --porcelain` must show no tracked change (untracked files do not count) and `git rev-parse --abbrev-ref HEAD` must print the parent. Either one fails: stop and report, and never check out, stash or reset to make it pass.
   Before deleting gate outputs, open the full cleanup procedure in `tidy.md` with a file-reading tool; shell reads of skill files may be sandboxed. If it cannot be read or completed, leave the files. Typed approval alone never replaces validation.
   Mode: `commit_history` in `acta config show`. `tidy` (default): follow [tidy.md](tidy.md), not steps 5-7. `full`: steps 5-7.
5. From the main checkout: `git merge --no-ff <branch> -m "<what landed>. Verified at merge: <gate numbers>"`. Capture the main checkout path first: cleanup runs from outside the worktree. A conflict: resolve it hunk by hunk by the intent of each side; never `--abort`, never drop a side. Cannot resolve it: stop, keep the worktree, report. Right after the merge, run `acta id --fix-duplicates` in the main checkout: two branches can take the same number, and this gives the later one the next free number (the hash never changes). Put each renumber line it prints in the landing report. Then run `acta land-notes <plan hash>` once: in one commit it writes `fixed_in` on the bugs the plan fixes, ticks the debt in its `closes:` and writes the dates of the plan's spec. Use the hash, since a renumber can change the id.
6. Run the gates again on the merge result, unless `git rev-parse HEAD^{tree}` prints the same tree as `git rev-parse <branch>^{tree}`: then the merge holds exactly the files the gates just passed, so write "tree same as branch, gates reused" in the report instead. Trees differ: list what changed with `git diff --name-only <branch> HEAD`; when every path sits under the planning root (the folder above `plans/` in the path `acta show <plan id> --path` prints), reuse the branch gates and write "only planning files differ, gates reused" in the report instead. Any other path, including Markdown outside the planning root, re-runs the gates: the full suite as before, and a slow gate only when the paths the parent gained since the branch left it (`git diff --name-only <merge base> <parent before the merge>`) call for it. Apply every slow-gate selection rule from step 1 to those paths: forcing paths also trigger a run, and a line with no run paths runs whole. Use the form those paths call for. Otherwise write "slow gate reused: the parent brought no path it covers" in the report, with the parent commits you read (`git log --oneline <merge base>..<parent before the merge>`). In doubt about what the parent gained, the gate runs. A plan with no `**Gates:**` line: a slow gate the repo's docs name runs again whole. Red: say so plainly and leave the merge for the user.
7. Clean up, only after the final plan check passes: `git worktree remove <path>`, then `git branch -d <branch>`. `-d` refusing means not fully merged: stop and report. Never delete a worktree that has uncommitted work. `git worktree remove` refusing: stop, show `git status --porcelain --ignored` of that worktree, and give the user two choices: the user removes the files, or stop the land and keep the worktree and branch. Add no commit here: the parent already moved, so it would never get it. Never `--force`, never delete the files first.
8. Never `git push` unless the user asks, never amend a commit, and never force anything. Last, from the main checkout, run `acta worktrees --clean`: it removes clean worktrees no open plan owns, never by force. Report each one it removed and each dirty one it kept.

### Final plan check

Run it in the main checkout after the parent moves, before any cleanup or success claim. Use the plan hash from `acta show <plan id> --json`, never the id.

1. `acta show <plan hash> --json`: `status` is `done`, and `progress.done` equals `progress.total`, above 0.
2. For each task N, `acta commits <plan hash> N --json` lists commits from every branch, so it alone proves nothing: one of them must pass `git merge-base --is-ancestor <sha> HEAD` there.

A failed check: stop, even when the tree is the same. Keep the worktree and branch, report the mismatch with the output, and claim no success. Never force a status, tick a box without proof, or redo work to hide it.

Report after landing: the merge sha first, the review level line from the plan's `## Review notes`, copied as it was written, the gate numbers you ran after the merge, for each slow gate the form that ran and why, or why it did not run, what was cleaned up, each flake with its bug, the wiki pages the branch added or changed (`git diff --name-only <merge-sha>^1 <merge-sha> -- .acta/wiki/`, or none), one next action, and, when there are debt items in the plan's `closes:`, the debt file and how many items it closed.

For retained temporary gate worktrees, the final cleanup next action follows the report decision in tidy.md. Incomplete or invalid proof means retain the worktree, without a deletion offer, approval question or suggestion that the user force-remove it.

### Understanding, after the land

Add this part only when `review_understanding` is true in the acta settings of the session note or in `acta config show`. A missing value is false. It is its own switch: quick, standard, deep, a floor or the deep lens never turns it on. A successful land adds it; a blocked or failed land does not.

- Keep every report part above. After the parent has moved (merge or tidy), read the landed files from the parent checkout, or with `git show <landed-sha>:<path>` from any folder. A read of the old worktree's files, or any read before the merge, does not count. Explain that final code (fixes, polish and the merge included), not the review notes or an earlier diff.
- Explain in a few lines: the main behavior, with one concrete example; the most important risk or boundary the code shows (if it shows none, say so and never invent one); what the tests do and do not prove.
- It asks the user no question and offers no quiz. The report still ends with the next action.

When `review_understanding` is false or missing, write the plain report: no explanation, no offer.
