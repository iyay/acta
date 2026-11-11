---
id: BUG-0086
hash: efes46s
priority: low
---
# TestTickWarningChangesNothingElse fails when the clock passes a second

## Symptom
`scripts/test ./internal/cli` failed once with `tick_test.go:548: plan files differ`. The only difference was the `started` and `finished` stamps, 06:40:58 against 06:40:59. Run alone, the test passed.

## Root cause
`internal/cli/tick_test.go:535-549` runs `acta tick --all` in two repos one after the other and compares the plan files byte for byte. Tick writes the current time, to the second, into `started` and `finished`. When the two runs fall on both sides of a second boundary, the files differ even though the warning changed nothing.

## Repro
Run `scripts/test ./internal/cli` a few times on a loaded machine. Or add a one-second sleep between the two `runTick` calls: the test fails every time.

## Found in
Branch telemetry-core (PLN-0228), review round 1, standards reviewer and the orchestrator's task 3 run.
