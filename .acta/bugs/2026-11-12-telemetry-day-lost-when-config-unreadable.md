---
id: BUG-0087
hash: qdjjfv1
priority: medium
---
# A telemetry day is lost when the user config cannot be read

## Symptom
On a release build, a user whose `~/.acta/config.yaml` fails to parse or validate sends days with `settings: {}`. The collector (walkie SPC-0194) needs all four setting keys, so it answers `invalid_input` with `retryable: false`, and acta deletes the day. That user's days never reach the Sheet.

## Root cause
`internal/cli/telemetry_run.go` `countRun` calls `SetSettings` only when `config.ResolveUser()` returns no error. `internal/telemetry/store.go` `newDay` starts `Settings` as an empty map, so a day where every resolve failed keeps it empty.

## Repro
Release build with an endpoint, a `~/.acta/config.yaml` that fails validation (for example `plan_depth: sideways`), run any command, then `acta telemetry status`: the body shows `"settings":{}`.

## Found in
Review round 1 of walkie PLN-0285 (collector acta route), spec axis, by reading acta main 8dca8de.
