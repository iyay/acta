---
created: "2026-09-30"
parent: bugs/2026-09-30-omp-build-runs-dispatch-executor
id: SPC-0043
hash: bybee7i
---
# omp Runs Subagents, and `acta voice` Becomes `acta config`

Status: approved by the user in chat on 2026-09-30 (Bounded). Fixes bug `.acta/bugs/2026-09-30-omp-build-runs-dispatch-executor.md`.

## Why

1. `build_executor: dispatch` is one value for every harness. An omp session reads it, follows `acta:dispatch` and opens yet another omp tab. That also hits an omp agent that got a dispatch brief. Dispatch is meant for harnesses other than omp. Inside omp, build should run its own `agent("task")` subagents.
2. The `acta voice` command now holds more than voice (executor, subagent models, theme). The user wants it called `acta config`, the same name as `~/.acta/config.yaml`.

## Design

### omp never dispatches

- `plugin/skills/build/SKILL.md`: the executor table and the `acta voice show` rule say that on omp, `dispatch` runs as `subagent` (`agent("task")`), with no question asked. Dispatch is only for harnesses other than omp.
- `plugin/skills/dispatch/SKILL.md`: a guard at the top. Running in omp: stop and go back to `acta:build` with the `subagent` executor.
- No harness detection in Go. The agent knows which harness it runs in, and the build skill already splits "Claude Code: ... omp: ...".

### Rename the command

- `internal/cli`: the `voice` case becomes `config`. The command file and its usage and error strings say `acta config ...`. `acta voice` is removed with no alias: it fails as an unknown command.
- Package `internal/voice` keeps its name. `internal/config` already exists and finds the planning root. Only the command changes.
- Every text that says `acta voice` changes to `acta config`: skills, `plugin/hooks/default-rules.md`, `plugin/README.md`, `plugin/evals/FACTS.md`, doctor messages, hook text (the `acta voice: reply in ...` line becomes `acta config: ...`), and the tests that check them.
- Planning history under `.acta/` is not rewritten.

## Testing

Failing tests first:

1. `internal/plugincheck`: the build skill states the omp rule; the dispatch skill has the omp guard.
2. `internal/plugincheck`: no file under `plugin/` contains `acta voice`.
3. `internal/cli`: `acta config show` and `acta config set` work as `acta voice` did; `acta voice` exits non-zero.
4. `internal/hook`: the prompt line starts with `acta config:`.

Gate: `scripts/test --full`, `go vet ./...`, `gofmt -l .`.

## After landing

Run `go install ./cmd/acta`, then restart every open Claude Code and omp session. Cached skill text still calls `acta voice` until then.
