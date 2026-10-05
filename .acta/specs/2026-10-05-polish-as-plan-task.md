---
parent: scratch/2026-10-05-polish-as-plan-task
id: SPC-0077
created: "2026-10-05 13:48:17"
hash: exdvp2b
started: "2026-10-05 13:55:00"
---
# A review polish is a task in the plan

Status: Bounded, approved by the user in chat on 2026-10-05. Also fixes BUG-0028.

Why: a polish today is one commit with no task. While it runs the plan shows every task done (5/5), so the TUI says nothing is left. A dispatched omp sees every task ticked, says the goal is done and makes no polish commit (BUG-0028).

Design:

1. Review skill, `## After a CLEAN round`: with `[fix]` NOTEs, the orchestrator first appends to the plan, in the worktree:
   ```
   ## Polish

   ### Task N: Review polish

   **verify:** every NOTE below is applied, and nothing else changes.

   - [ ] <one box per [fix] NOTE>
   - [ ] Commit: `polish: review notes for <plan id>`
   ```
   N is the next task number. The polish still counts as no review round; its revert and debt rules stay as they are.
2. `acta dispatch send --round polish` takes its tasks from the `## Polish` section, the way a fix round takes them from its `## Fix round` section, and names them in the brief and the checkpoint. A plan with no `## Polish` section refuses with an error that says to add it. The note stays optional. The first dispatch (no round) stops at `## Polish` as it stops at `## Fix round`.
3. `acta reply-back` already refuses while a task is open, so an unticked polish task keeps the recipient working.
4. `plugin/skills/build/dispatch.md` says the same in one line.
5. The last task adds 1 to the patch version in the three plugin files.

Tests: brief for `--round polish` names the polish task; a plan with no `## Polish` refuses; the first dispatch leaves the polish task out; the plan shows the polish task as open until ticked.
