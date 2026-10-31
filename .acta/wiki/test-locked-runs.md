---
type: Reference
title: Run tests the locked way
description: scripts/test serializes full runs with run-one, pre-tool blocks bare go test, testguard kills orphans
paths: [scripts/test, cmd/flakerun/, internal/flakerun/]
---

`scripts/test --full` runs under the machine-wide `acta run-one` lock, so one full suite runs at a time. It goes through `go run ./cmd/flakerun`, which reruns each failed test alone up to 2 times and prints `FLAKY: <package path>.<TestName> (failed <k> of <n>)` for one that passes on a rerun. The run exits 0 only when every failed test passed on a rerun; a package whose test binary crashed (panic, timeout or os.Exit) reruns nothing and fails the run. The pre-tool hook blocks bare `go test` in this repo once the PATH acta is rebuilt; every test package must call `testguard.Watch()` or plugincheck fails. Per-package runs and `-short` keep contention down. Slow shared suites come mostly from CPU fights with git-spawning packages, not from missing `t.Parallel`.
