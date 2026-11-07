---
parent: scratch/2026-11-07-stray-agent-worktrees
id: SPC-0189
created: "2026-11-07 16:47:20"
hash: c8dyztl
started: "2026-11-07 17:43:59"
finished: "2026-11-07 19:04:18"
---
# Stray worktrees

## Problem

Agents make git worktrees outside the build flow, for head checks, pre-merge checks or test runs. Nothing removes them. In be-pmis three were left behind (`be-pmis-headchk`, `be-pmis-premerge`, `be-pmis-ps-export`). The agent then tried `git worktree remove --force`, the pre-tool hook blocked it, and the user had to run the commands by hand. Users forget, and the worktrees pile up as junk.

acta itself makes one worktree per plan in `acta:build` and removes it in `acta:land` with a plain `git worktree remove`. The hook allows that.

Facts checked in a temp repo:
- `git worktree remove` without `--force` succeeds when the worktree holds only gitignored files, like `node_modules`.
- It fails with exit 128 when the worktree has untracked or modified files.
- It succeeds on a clean detached worktree.

## Goal

No worktree is left behind as junk, and acta still never forces a delete.

## Out of scope

- The `--force` block in the pre-tool hook stays as it is. A forced remove is always the user's to run.
- No session-end hook.
- `acta cleanup` stays about planning items only.

## Design

### Stray worktree

A stray worktree is a linked worktree of the repo that no open plan owns. These are never strays and are never touched:
- the main checkout;
- the worktree the command runs in;
- a worktree whose branch has commits the main checkout's HEAD does not hold (`git rev-list --count HEAD..<branch>` from the main checkout prints more than 0). A plan with every box ticked reads as done before it lands, so plan status alone cannot protect its worktree;
- a worktree whose branch name equals the slug of a plan that is neither done nor dropped. This covers a fresh build worktree that has no commits yet;
- a locked worktree (`git worktree lock`);
- a worktree with activity in the last 24 hours, meaning its `HEAD`, `index` or `logs/HEAD` in the worktree's git dir changed then. This keeps a live agent in another tab safe.

The protected ones are listed with their reason, so the user sees them.

The 24-hour check reads those file times before acta runs any git command in that worktree. Every `git status` that acta runs uses `--no-optional-locks`, because a plain `git status` rewrites the `index` file and would reset the clock.

A stray is clean when `git --no-optional-locks status --porcelain` in it prints nothing. Gitignored files do not count. Otherwise it is dirty.

A worktree whose folder is gone (git marks it `prunable`) is a missing stray. It is checked before the rules above, since its folder holds nothing left to lose. A lock still comes first. A detached one is still checked for commits that no branch, tag or remote holds: such a commit lives only in that worktree's entry, so the worktree stays protected. `--clean` removes its entry with plain `git worktree remove <path>` and then tries `git branch -d` like any other stray. A missing folder never makes the command fail.

Worktrees a harness makes inside the repo, like `.claude/worktrees/<name>`, follow the same rules.

### `acta worktrees`

- With no flag it only lists. Each stray shows its path, its branch (or detached) and clean or dirty. A dirty one also shows `git status --short`, at most 10 lines, then `... and N more`.
- `--clean` removes each clean stray with `git worktree remove <path>`, never `--force`. Then it runs `git branch -d <branch>` when the stray had a branch. When `-d` refuses because the branch is not merged, the branch stays and the output says so. Dirty strays are listed, never removed.
- `--json` prints one object per line for every linked worktree, protected ones included. The main checkout is left out. Fields: `path`, `branch` (empty when detached), `state`, `reason`, `branch_kept`, `status` (the `git status --short` lines, at most 10). `state` is one of:
  - `protected`: a rule above keeps it; `reason` names the rule.
  - `clean`: a clean stray, listed only.
  - `dirty`: a dirty stray; `status` holds its lines.
  - `missing`: a missing stray, listed only.
  - `removed`: `--clean` removed it. `branch_kept` is true when `git branch -d` refused, and `reason` holds the refusal.
  - `kept`: `--clean` tried to remove it and git refused; `reason` holds git's error text.
- Exit 0 when the listing or the clean ran, even if some items were kept or a remove failed. Exit non-zero only when git cannot list the worktrees or the folder is not a git repo.
- Help text and the man page name the command.

### Wiring

- `tidy.md` step 3: the detached temp worktree made to run the gates on the tidy tip is removed with plain `git worktree remove <path>` right after the gate run, red or green. Today nothing removes it; this repo still holds one (`acta-tidy-destructive-skill`).
- `acta:land` step 7, and step 7 of `tidy.md`: after the plan's own worktree and branch are gone, run `acta worktrees --clean` from the main checkout. The land report lists what it removed and each dirty stray it kept.
- `acta doctor`: a new `worktrees` check. It warns when any stray exists, names the count, and gives `acta worktrees --clean` as the fix. No strays gives OK.
- Core rule 10, in the session-start rules (`coreRules` in `internal/hook/hook.go` and `plugin/hooks/default-rules.md`, kept the same):

  > 10. A worktree you make outside acta:build is yours to remove before you report: run git worktree remove without --force. When it refuses, show git status --short of that worktree (10 lines at most) and ask the user: commit, discard, or they run the force themselves.

### Testing

- Go tests on temp repos, one case each: fresh plan worktree matched by slug, branch with unmerged commits, fully ticked plan waiting to land, dropped plan with no commits, locked, active within 24 hours, clean stray, dirty stray, detached stray, missing folder, harness worktree inside the repo, only gitignored files, a remove that git refuses. Each case checks the `state` in `--json` and whether `--clean` removed it.
- A test that a plain listing does not change any worktree's `index` time, so a `--clean` right after a listing still removes an idle clean stray.
- A test that `--clean` never passes `--force` or `-f` to git.
- A test that `acta doctor` warns with strays and is OK without.
- A plugin eval for land step 7: the agent runs `acta worktrees --clean` after removing its own worktree and puts the result in the report.
- A plugin eval for tidy step 3: the agent removes the temp gate worktree after the gate run.

## Wiki

Approved: add `stray worktree` to `glossary.md`: "a linked worktree no open plan owns, not locked and idle for 24 hours; `acta worktrees --clean` removes it when clean."
