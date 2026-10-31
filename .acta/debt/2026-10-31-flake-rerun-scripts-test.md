---
id: DBT-0149
hash: rfo3hi1
parent: plans/2026-10-31-flake-rerun-scripts-test
---
# Review NOTEs: Flake rerun in scripts/test implementation plan

- [ ] (medium) run-one sends SIGTERM to go run ./cmd/flakerun; go run dies on SIGTERM, so flakerun is orphaned with no testguard and keeps running the suite and its reruns after run-one drops the lock (cmd/flakerun/main.go, internal/cli/runone.go). Found by reading the source, not run.
- [ ] (medium) A -failfast run is still a false green in internal/flakerun: later tests in the package get no run event, so a flaky TestA that passes on rerun prints FLAKY and exits 0 though a real red TestB never ran. Nothing passes -failfast today; refuse it, or count tests run against tests in the package.
