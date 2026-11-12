---
id: SPC-0210
created: "2026-11-13 11:29:55"
hash: g4glcwb
---
# Temporary gate operations use the owning repository

## Why

BUG-0072 tracks intermittent failure of `land-tidy-temp-worktree`. During PLN-0242 landing, this case failed its gate and its one allowed known-flake retry. The retry created and tested a detached gate worktree, attempted removal from the outer workspace, received `not a working tree`, moved the parent, then removed the same path successfully from the owning repository. The required removal-before-parent order failed.

A controlled nested-repository clone reproduces exit 128 from the outer repository and exit 0 from the owning repository for the same registered path. This explains the retry, not every historical failure. The initial gate trace omitted detached-worktree creation and remains a separate unverified mechanism.

## Design

This is a Bounded correction to the existing tidy landing flow. The user approved this design on WhatsApp on 2026-11-13, in reaction to message 3EB05D2617566638D01854.

Capture the absolute owning repository checkout before creating the detached gate worktree. Run temporary-worktree creation, listing and removal explicitly through that repository, using `git -C` or an equivalent checked working directory. Keep the tested worktree path separate from the repository that owns its registration. Verify removal from the owning repository before moving the parent when the temporary worktree is clean and removable. An error from the wrong repository cannot establish successful cleanup or a file-based removal refusal.

Preserve existing red-gate stops, genuine refusal handling, cleanup permissions, final plan checks and parent-movement rules. Add no force, recursive deletion, new CLI or configuration. BUG-0092's ignored-only initial-removal problem is outside this design. Do not weaken the eval's requirement that successful temporary-worktree removal precedes parent movement.

Strengthen the existing temporary-worktree eval with an outer repository and a distinct owning repository. Observe actual registered worktrees, command exit statuses and order. Positive controls accept owner-scoped removal before parent movement; negative controls reject wrong-repository removal, removal only after movement, a different worktree's removal and force. Reuse current runner and grader types.

## Testing and sequence

Watch focused plugincheck tests fail before changing the instructions. Validate changed graders with positive and negative samples. Compare at least 10 old-text and 10 new-text runs under identical corrected fixtures for `land-tidy-temp-worktree` and neighboring `land-tidy-temp-refused`; retain raw evidence and report residual failures. Adjust only necessary text caps.

Build inline in its own worktree from main. Land this correction first, then resume PLN-0242 while preserving both approved changes in overlapping land files. Run the gates required for each resulting tree. Continue PLN-0237 through PLN-0241 afterward without repeating their approvals. Never push. Do not attach documents unless requested.

Primary files are `plugin/skills/land/tidy.md`, the `land-tidy-temp-worktree` eval and focused plugincheck checks. Follow the existing land/tidy and eval wiki pages. The plan uses `parent: bugs/2026-11-07-flaky-eval-land-tidy-temp-worktree`. Record `fixed_in` only after landing; do not claim all historical flakiness is eliminated.
