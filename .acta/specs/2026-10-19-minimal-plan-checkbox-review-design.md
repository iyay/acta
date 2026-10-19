---
id: SPC-0164
created: "2026-10-19 20:26:30"
hash: noye69p
---
# Check minimal task checkboxes before hand-off

## Problem

BUG-0050 exposed a gap in minimal-plan self-review. Slice requires Red,
Change and Commit checkboxes, but its minimal review checks only spec
coverage and property-shaped verify lines. Both split plans were written
without checkboxes, so ticking their tasks failed.

## Design

Update the minimal-plan self-review in `plugin/skills/slice/SKILL.md` to
check every task in every generated plan for the three required checkbox
steps: Red, Change and Commit. Missing boxes must be corrected before
the plan is handed to the user. Preserve the existing coverage and verify
checks, approval gates and full-plan format.

Keep the wording short and consistent with the existing minimal format.
Do not add a CLI validator or change tick behavior.

## Verification

Add a failing regression check in `internal/plugincheck/skill_slice_test.go`
for the minimal self-review requirement. Verify that removing that
requirement makes the check fail. Run the relevant plugin checks,
including skill size limits, and the required full gates before landing.

The implementation plan must use
`parent: bugs/2026-10-19-tick-all-fails-on-minimal-plan`.
