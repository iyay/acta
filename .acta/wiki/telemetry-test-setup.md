---
type: Reference
title: Telemetry is silent unless a test turns it on
description: telemetry.Endpoint is empty outside release builds, so tests count nothing unless they set Endpoint and HOME
paths: [internal/telemetry/, internal/cli/telemetry.go, internal/cli/telemetry_run.go]
---

`telemetry.Endpoint` is set only by the release `-X` line (SPC-0201). Every `go test`, `go build`, `go install` and eval binary has it empty, and then `Store.Enabled` is false: nothing is counted, stored or sent. So most tests never see telemetry, and evals never send.

A test that wants counts sets three things, the way `telemetryHome` in `internal/cli/telemetry_test.go` does:

- `telemetry.Endpoint` to any non-empty URL, put back in `t.Cleanup`.
- `HOME` to `t.TempDir()`, since the store lives in `~/.acta/telemetry/`.
- `PM_VOICE_FILE` into that temp home, since `countRun` reads the settings snapshot from the user config.

Also clear `DO_NOT_TRACK` and `ACTA_TELEMETRY`; either one turns counting off.

`cli.Run` counts itself before it prints, so a test that reads today's counts after a run sees that run too (`acta telemetry status` counts as `telemetry`).

Flush tests use `httptest.NewTLSServer` and `newClient(ts.Client().Transport)`. Never point a test at a real address.
