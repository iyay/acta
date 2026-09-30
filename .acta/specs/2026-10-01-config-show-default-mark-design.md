---
id: SPC-0058
created: "2026-10-01 06:45:44"
hash: mq4pa6w
started: "2026-10-01 06:50:22"
finished: "2026-10-01 06:51:50"
---
# config show marks plan_depth when it is only the default

Status: design approved by the user in chat on 2026-10-01. Bounded: `acta config show` (`internal/cli/config_cmd.go`) and the setup skill already exist.

## Why

`acta config show` prints `plan_depth: full` when no file sets it. The setup skill asks only for what `config show` says is not set yet, so it cannot tell "unset" from "set to full" and skips the plan depth question. This happened on 2026-10-01.

## Design

- The text output still always prints the `plan_depth` line, because the plan skill reads it. When neither the user file nor `.acta.yaml` sets it, the line reads `plan_depth: full (default)`. This works the same way as the ` (repo)` mark.
- A value set to `full` in either file prints with no `(default)` mark. A repo value keeps ` (repo)`.
- The JSON output does not change: `"plan_depth": "full"`. No skill reads the JSON.
- `plugin/skills/setup/SKILL.md`, First run step 2: add that a value marked `(default)` is not set yet, so setup asks for it.
- The other fields stay as they are. They are already left out when unset.

## Testing

- Update `internal/cli/config_repo_test.go` (the test "With nothing set anywhere") to expect `plan_depth: full (default)`.
- New test: the global file sets `plan_depth: full`, and show prints `plan_depth: full` with no `(default)` mark.
- `internal/plugincheck` stays green after the setup skill edit.
