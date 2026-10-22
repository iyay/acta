---
created: "2026-10-22 23:43:49"
id: SPC-0175
hash: ymhb9jb
---
# Verify plan completion through build and land

## Problem
PLN-0176 reached the parent branch while its board progress remained 1/3. BUG-0050 records missing task checkboxes, and BUG-0051 records lost task trailers during tidy. Those fixes do not establish that an agent verifies the final plan in the parent checkout before reporting success.

The current build skill already requires ticks and a progress check before review. The land skill checks progress before merging. Strengthen these existing checks and verify the final result through behavior evals.

## Scope
This is a Bounded change to the build and land skills, their focused contract tests, and behavior evals. Keep CLI status derivation and tidy implementation unchanged. Do not close or rewrite old plans as part of this change.

## Behavior
- Build must read the result of each progress-writing command. A failed tick blocks further task work, review, and land until the error is resolved. Missing checkboxes are a plan-format error, not permission to skip progress recording or force a status.
- After committing and ticking a task, build must read that task back and verify its completion. Before review, the plan must contain tasks, every task must be complete, and the plan progress must match. A commit alone is not completion evidence.
- Both full-history and tidy landing must verify the plan by its stable hash from the parent checkout after the parent moves. The plan must be done, every task must be complete, and each task must retain its linked implementation commit evidence. Use existing CLI commands for these checks.
- A failed final check blocks worktree or branch cleanup and any completion claim. Report the mismatch and preserve the evidence; do not force status, tick unchecked work without proof, or redo implementation to hide the mismatch.
- Existing test, review, wiki, clean-tree, and merge gates remain required. Keep the new instructions shared between landing modes where possible.

## Validation
Use focused Go contract tests to protect the required ordering and stop conditions. Use behavior evals with disposable repositories to prove that an agent handles missing checkboxes and a failed tick without proceeding, detects incomplete task progress before review, and verifies final progress and task commit links for both full-history and tidy landing. Include a final-parent mismatch that must preserve the worktree and must not produce a success report. Grade repository state and executed checks rather than the presence of words in a response.

Follow TDD in an isolated worktree. Run focused checks during implementation, then the required full suite and skill eval gates before landing. Review and land follow the configured workflow.

## Boundaries
Skills guide agent behavior; they do not add a CLI enforcement mechanism. The second repository shown in the screenshot has not been inspected, so this change does not claim to establish its historical cause. Existing plans remain evidence and are not repaired by this work.
