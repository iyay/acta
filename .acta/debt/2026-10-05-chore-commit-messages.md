---
id: DBT-0077
hash: luwk4dc
parent: plans/2026-10-05-chore-commit-messages
---
# Review NOTEs: chore commit messages Implementation Plan

- [x] (high) newestRound in internal/board/state.go passes a commit hash to gitc.FirstSeen, which wants a file path, so the order check never works: on a branch with both `chore(plan): tick fix round` and `acta: tick fix round` subjects, acta state always shows the new-subject round even when the old one is newer. Rank by place in rev-list instead, and fix the doc comment.
- [x] (low) TestRunningRoundPrefersNewSubjectWithoutOrder passes only because of the FirstSeen mix-up above; once ordering works it will fail, since its old-subject commit is the newer one.
- [ ] (low) The TUI status line still shows `acta: <id> ...` after a write (internal/tui/model.go 892, 933, 1087), while commits now read chore(...).
