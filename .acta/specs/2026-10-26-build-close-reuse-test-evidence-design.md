---
parent: bugs/2026-10-22-build-close-reruns-fast-suite
id: SPC-0178
created: "2026-10-26 06:22:13"
hash: q32azyk
---
# Reuse task test evidence during build close

## Problem

BUG-0054 records build close rerunning the repository-wide fast suite. The
Close section asks the agent to show outstanding type checks with tests each
task ran. That wording can prompt another test run to produce fresh output,
despite the adjacent no-second-suite rule.

During the Cold read landing gate, `build-close-no-suite` failed on the
candidate and on the unchanged parent. Both traces executed `scripts/test`
after `scripts/typecheck`. The existing grader found `broad-suite-ran`.
Cold read does not change build close or this eval.

## Scope and behavior

This Bounded change clarifies the existing Close section in
`plugin/skills/build/SKILL.md`.

- Close reads the committed task results and reports the tests already run.
  It does not rerun completed checks to obtain output for the report.
- Close runs outstanding type checks. If later work changed behavior, run
  only the focused check for that change, as the current workflow requires.
- Close does not run a repository-wide fast or full suite. The full suite
  remains a landing gate.
- Preserve wiki review, task-count and completion checks, implementation
  review, dispatch close, and every existing error stop.

Keep the CLI, commands, settings, test runner, grader assertions and thresholds
unchanged. The user approved a fixture revision after three wording variants:
store the broad-suite and typecheck evidence markers under `.git/`, then update
their grader paths and the existing fixture path test. A successful typecheck
was followed by deletion of its root marker as a stray file; that lost evidence
made the grader report a skipped check. Preserve the same execution evidence
without exposing marker files as workspace cleanup targets.

Do not include the separate intermittent `build-tick-failure`
case, which passed its baseline and candidate focused rechecks.

## Verification

Use the existing `build-close-no-suite` behavior eval as the regression test.
Record repeated baseline and candidate results under the same client and
model settings, with the same revised fixture in both arms; run five candidate
trials. Every candidate trial must leave `.git/broad-suite-ran` absent, create
`.git/typecheck-ran`, and reach implementation
review without starting it. Keep all existing graders and their thresholds.

Run the focused build skill and budget checks, then the normal review and
landing gates. Stay within existing skill budgets. Trial results measure this
scenario; they do not establish that an LLM can never violate an instruction.

## Relationship to Cold read

Use a separate plan and implementation worktree. The Cold read branch stays
intact and unlanded until its gate is resolved. No implementation is approved
by this document alone.
