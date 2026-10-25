---
parent: debt/2026-10-23-plan-completion-skill-gates
closes: DBT-0142.02
depth: minimal
id: SPC-0179
created: "2026-10-26 07:11:03"
hash: rtcxm4k
---
# Make the build close eval trustworthy

## Goal

Resolve the red `build-close-no-suite` eval without weakening its checks. The eval must prove that build close runs the outstanding type check, skips the repository-wide fast suite, and hands off to review.

## Scope

Run the eval from the repository checkout on main and on the current branch, following the eval runbook. Keep traces and record which tools the agent called and which skill files it read. Compare the scaffold, grader, runner, and build instructions to identify the cause of each failure.

Change only the fixture or build instructions needed to fix a cause proven by those runs. Keep the type-check marker and no-suite marker assertions. Add or adjust focused contract coverage only when it captures the proven failure. Run the eval repeatedly from a supported checkout location to confirm the result.

Do not change unrelated build behavior, loosen a grader, or run the full repository suite during eval diagnosis. The close flow must still stop before review. Landing retains responsibility for the full suite.

## Acceptance

- The eval passes on the repaired branch in repeated runs.
- Its type-check marker exists and its broad-suite marker does not.
- The final response reaches the review handoff without starting review.
- The kept trace shows that the agent did not invoke the review skill.
- Any skill change stays within plugincheck size limits.
