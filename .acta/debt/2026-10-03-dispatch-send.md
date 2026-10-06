---
id: DBT-0071
hash: ctm2g5f
parent: plans/2026-10-03-dispatch-send
started: "2026-10-06 09:30:59"
---
# Review NOTEs: Dispatch Send and Close Implementation Plan

- [ ] (high) omp folds each todo phase to 8 rows, so a round with more than 8 tasks in one phase shows only some ids and checkpoint in internal/cli/dispatch_herdr.go reports a false drift (exit 4). dispatch.md now says read the pane first, but the verdict itself is wrong. Read the card's closed/total count or judge only rows on screen.
- [ ] (high) The ascii preset fix only covers an unframed card. With tasks, omp frames the header as `+--- [x] Todo 3 tasks ---+`, which todoCardHeader does not match, so ascii users still always get unconfirmed. A regex that covers both: ^[^\p{L}\p{N}\n]*(?:\[x\][^\p{L}\p{N}\n]*)?Todo(?:\s|$). TestHerdrCheckpointAsciiHeader tests a form omp never draws when tasks exist.
- [ ] (medium) On a reused tab (fix round, polish) the last round's Todo card can still be in the 60 recent lines, so a slow agent gives a false drift on the new task ids. Read only the pane text after the newest goal.
- [ ] (medium) deliverGoal on a reused tab passes when a goal mark from an older goal is still on screen, so it can report success before the new goal is set. Require the mark to be new, for example by comparing with the read taken before the prompt.
- [ ] (medium) acta dispatch send writes .claude/dispatch/<slug>-brief.md without checking that .claude/ is git-ignored. In a repo that does not ignore it, a recipient's git add -A can commit the brief. The old herdr-delivery.md had this check.
- [ ] (low) findOrMakeTab reuses any agent with the slug's name without checking that it runs omp in this worktree, so two repos with the same branch name would share one agent.
