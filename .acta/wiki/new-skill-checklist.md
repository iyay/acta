---
type: Runbook
title: Adding a skill or a command touches more than its folder
description: A new skill or command must join hook.Skills, telemetry lists and plugincheck rows; regen default-rules.md
paths: [plugin/skills/, internal/hook/hook.go, internal/telemetry/store.go, internal/plugincheck/]
---

A new skill folder under `plugin/skills/` fails Go tests until each of these lists knows it:

- `hook.Skills` in `internal/hook/hook.go`. The session-start text prints this list, so `plugin/hooks/default-rules.md` changes too: regenerate it with `scripts/test ./internal/hook -run TestDefaultRulesFile -update`. That file sits in `plugin/hooks/`, so the eval gate then runs whole at land, not only `--tag=<skill>`.
- `skillKeys` in `internal/telemetry/store.go`.
- `fileCaps` and `descriptionCaps` in `internal/plugincheck/budget_test.go`.
- A row in `evalCases` in `internal/plugincheck/evals_test.go`, and an eval case whose files hold `# guards: <phrase>` with the exact phrase of that row.

A new CLI command needs its name in `commandKeys` in `internal/telemetry/store.go`, or `TestEveryHelpCommandIsAKnownTelemetryName` fails once `help.go` lists it.
