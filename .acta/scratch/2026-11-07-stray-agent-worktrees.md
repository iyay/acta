---
id: SCR-0090
hash: ly1zg25
title: Agent-made worktrees are left behind as junk
status: brainstorming
created: "2026-11-07 16:40:37"
schema: "1"
started: "2026-11-07 16:42:23"
finished: "2026-11-07 16:47:20"
---
# Agent-made worktrees are left behind as junk

## Words

### 2026-11-07

ini kenapa diblock ya? [image /Users/iyay/.local/state/walkie/media/3AB133B5021D3C7CBE94.jpg]

loh kok worktree dihapus gaboleh?

kan acta sendiri yg nyuruh kerja pake worktree

mau.. intinya kalo isseunya kayak di gambar jadinya sampah. user bisa aja lupa remove

## Context

Seen 2026-11-07 in the be-pmis project. An agent there wanted to clean up three worktrees (../be-pmis-headchk, ../be-pmis-premerge, ../be-pmis-ps-export) and two branches with `git worktree remove --force` and `git branch -D`. The pre-tool hook blocked the `--force` command (internal/hook/destructive.go:150 blocks any git command with `--force`; :176 blocks `worktree remove -f`). Chat approval cannot lift the block, so the agent handed the commands to the user to run by hand.

The land step removes its own worktree with plain `git worktree remove` (plugin/skills/land/SKILL.md:67, tidy.md:13), which the hook allows. The leftovers here look like worktrees an agent made outside the build flow (head checks, pre-merge checks), so nothing cleans them up. The user may forget, and they pile up as junk.

## Log

### 2026-11-07

2026-11-07 brainstorm start (path: Architectural, user can override).
Facts found:
- acta itself makes only one worktree: build/SKILL.md:47 (`git worktree add ../$REPO-$SLUG`). Land removes it with plain `git worktree remove` (land/SKILL.md:67, tidy.md:13).
- The be-pmis leftovers (headchk, premerge, ps-export) come from agent checks outside the build flow. omp managed skills for be-pmis include branch-divergence-check and worktree-jest-run, which likely make such worktrees.
- Tested in a temp repo: `git worktree remove` without --force succeeds when only gitignored files exist (node_modules); fails with exit 128 on untracked or modified files; succeeds on a clean detached worktree.
- doctor has no worktree check today (internal/doctor/doctor.go).

### 2026-11-07

Round 1 answers (user: "setuju, yg penting gak ninggalin sampah" = accept all recommendations):
- Q1 fix lives in a skill rule plus an acta doctor check. No session-end hook.
- Q2 stray = any linked worktree that is not the worktree of an open plan.
- Q3 doctor --fix removes clean strays (plain remove, branch -d); dirty ones are only listed, never forced.
- Q4 the --force block in the pre-tool hook stays as it is.
- Q5 when plain remove fails, the agent shows git status --short (at most 10 lines) and offers commit, discard, or the user runs force.
Goal in the user's words: no junk left behind.

### 2026-11-07

Round 2 answers (user: "setuju" = accept all):
- Q6 land, after removing its own worktree, also removes clean strays and lists dirty ones in its report. acta doctor can still be run by hand.
- Q7 skip any worktree with activity in the last 24 hours, so a live agent in another tab is never hit. Fresh ones are the job of the agent rule.
- Q8 locked worktrees are always skipped and only listed. Harness-made worktrees inside the repo (like .claude/worktrees) count as strays under the same rules.

### 2026-11-07

Shared understanding confirmed ("swtuju"). Approach A chosen: a new `acta worktrees` command. No flag lists strays; `--clean` removes the clean ones. Land calls `--clean`; doctor calls the list for a warning. Rejected: B (doctor --fix also touches gitignore and wiki), C (skill text only, no tests, agents read it differently).

### 2026-11-07

Design approved ('oke'), all three sections: 1 acta worktrees behaviour, 2 wiring (land step 7, doctor check, core rule 10), 3 tests (Go temp repos, no --force, land eval).

### 2026-11-07

Cold read: 5 findings (done plan unprotected, git status resets idle clock, tidy temp worktree, missing folders, json shape). User approved all 5 fixes and the glossary term with a thumbs up; spec updated.

## Open questions

- Should acta track every worktree an agent makes (not only the build one) and clean it up at land or session end?
- Should `acta doctor` list stale worktrees and branches already merged into the parent?
- When plain remove fails because the tree is dirty, what should the agent show the user so the choice is quick?
