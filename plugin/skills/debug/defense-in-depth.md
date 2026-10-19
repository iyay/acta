# Checks After a Root-Cause Fix

Fix the cause first. Then ask whether the bad value could arrive again by
another path. Add a check only where it earns its place.

## Where a check belongs

- **Trust boundaries.** Data that comes from outside the code you control:
  user input, files, network replies, environment, other services. Check it
  where it enters, and fail with a clear message.
- **Invariants that must hold.** A rule whose break would lose data or do
  something unsafe, such as running `git init` outside a temp folder in a
  test. Guard it right before the dangerous step.
- **Nowhere else by default.** Internal code that only receives values you
  already checked does not need its own copy of the same check. Repeated
  checks hide which one matters and drift apart over time.

## Example

Bug: an empty `projectDir` made `git init` run in the source folder.

- Cause, fixed at the source: a test read `tempDir` before setup filled it.
  Now reading it early throws.
- Trust boundary: `Project.create()` rejects an empty or missing directory,
  since callers pass it in.
- Invariant guard: in tests, the function that runs `git init` refuses any
  folder outside the temp folder, because a mistake there damages the repo.

`WorkspaceManager`, between the two, gets no extra check: it only receives
the value `Project.create()` already checked.

## Test each check

Each check you add needs a test that sends the bad value and sees it
rejected. A check with no failing test is a guess.
