---
type: Reference
title: Run tests the locked way
description: scripts/test serializes full runs with run-one, pre-tool blocks bare go test, testguard kills orphans
paths: [scripts/test]
timestamp: 2026-10-05T15:28:00Z
---

`scripts/test --full` runs under the machine-wide `acta run-one` lock, so one full suite runs at a time. The pre-tool hook blocks bare `go test` in this repo once the PATH acta is rebuilt; every test package must call `testguard.Watch()` or plugincheck fails. Per-package runs and `-short` keep contention down. Slow shared suites come mostly from CPU fights with git-spawning packages, not from missing `t.Parallel`.
