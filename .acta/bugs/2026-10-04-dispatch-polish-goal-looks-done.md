---
id: BUG-0028
hash: b86e4y3
fixed_in: ab80e3c2773a19386bbe9c33aba8275a669cbc5e
finished: "2026-10-05 14:34:24"
---
# A polish dispatch is skipped: omp says the goal is done and makes no commit

## Symptom
`acta dispatch send --round polish` reaches omp, but omp re-audits the finished plan, reports "goal achieved" and makes no commit. The `[fix]` NOTEs in the brief are never applied. The orchestrator only learns this from an empty `git log <base>..HEAD`.

## Root cause
`internal/cli/dispatch_send.go:185` builds the `/goal` text from the plan title only, the same for every round. The brief's first line does say "Apply the review notes on ...", but the goal does not. A polish adds no task to the plan, so every task is already ticked, and omp takes the plan-title goal as met before it acts on the NOTE. A fix round is less exposed because it appends open `## Fix round` tasks to the plan.

## Repro
1. Build a plan through dispatch until every task is ticked and reply-back exits 0.
2. Run `acta dispatch send --plan <plan> --rules <rules> --round polish --note-file <notes>` in the worktree.
3. Wait for omp to go idle. `git log <base>..HEAD` is empty and the pane says the audit is done.

## Found in
main, while landing PLN-0080 (branch frame-probe, base 74f4882) on 2026-10-04: the first polish send got a 58 s re-audit and no commit. A plain herdr prompt naming the polish was needed to get the work done.
