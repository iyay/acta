---
id: DBT-0136
hash: g1vu153
parent: plans/2026-10-20-close-without-fast-suite
started: "2026-10-20 08:13:04"
finished: "2026-10-20 08:34:59"
---
# Review NOTEs: Close builds without another fast suite

- [x] (medium) The dispatch brief prints the plan whole **Tests:** line under GATES, fast command included (internal/cli/dispatch_brief.go), so a dispatched build can still run the broad suite before review. Decide whether the brief should name only each task focused tests.
- [x] (medium) The eval markers in plugin/evals/build-close-no-suite/scaffold.sh follow the caller cwd. A scripts/test run from a subfolder writes broad-suite-ran below the grader path, so the eval can pass while the close still ran the broad suite. Anchor both marker paths to the workspace root.
