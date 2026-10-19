---
created: "2026-10-20 05:29:39"
id: SPC-0166
hash: w848ote
---
# Start builds without a baseline test run

## Problem

Build setup runs tests before the first task. Slow suites delay implementation,
and agents sometimes run a completed suite again just to recover its summary.

## Design

Bounded workflow change approved in chat on 2026-10-20.

- After worktree setup, start the first task without running baseline tests.
  This applies even when no saved baseline exists or the saved result is old.
- Keep each task's failing regression test, focused verification, the close
  checks, review, and the full landing gates.
- Capture the full suite's output and exit status during its existing run.
  Read that log for summaries; never repeat a run only to recover its output.
- After every full suite PASS in the landing flow, save the latest passing
  baseline in `acta-baseline.md` inside the repository's absolute Git common
  directory. Keep its log beside it. Linked worktrees share this local record;
  it is not committed and survives worktree cleanup.
- Record the tested commit and tree, command, time, relevant runtime and
  dependency context, exit status, summary, and absolute log path. Publish the
  record atomically after the complete run succeeds. Preserve the last PASS
  and its log when another run fails; retain the failed run's output separately.
- The record is historical evidence, not permission to skip landing gates.
  A different tree or test context is not a current PASS. Existing merge and
  tidy gate-reuse rules stay unchanged. Reusing gates does not create a new
  full-suite result or relabel an old result as freshly tested.
- A missing or unreadable record does not trigger startup tests. Diagnose a
  possible pre-existing failure only when a task actually encounters one.

## Scope and verification

Change the build and landing skills, their focused plugin checks, the build
flow eval, and landing eval coverage for the saved result. Cover ordinary and
tidy landing through the shared landing preconditions. Update the wiki pages
that describe these flows. No new CLI command, dependency, cache engine, or
test-runner integration is needed.

Verify that an approved build reaches its first task without a baseline test
run, including absent and stale records. Verify that landing retains its gates,
publishes a passing record with readable output, preserves the last PASS after
failure, and reads existing output rather than running a suite for its summary.
Use focused checks during implementation and the normal full gates at landing.
