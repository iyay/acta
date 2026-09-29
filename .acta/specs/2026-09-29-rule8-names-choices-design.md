---
parent: bugs/2026-09-29-second-brainstorm-choices-not-offered
id: SPC-0024
hash: lzoio4u
---
# Session Rule 8 Names the Second-Brainstorm Choices

Status: approved by the user on 2026-09-29 (Bounded). Fixes bug `.acta/bugs/2026-09-29-second-brainstorm-choices-not-offered.md`.

## Why

- Asked for a second big change in one session, the agent answers from session-start rule 8 alone and never loads `acta:brainstorm`. A failing eval trace has zero tool calls.
- Rule 8 says only "the user picks how to open it". It names no choices, so the agent makes some up. Eval case `second-brainstorm-choices` fails most runs, on main too.
- Real users hit the same path, not only the eval.

## Design

1. `internal/hook/hook.go`, `coreRules` rule 8: keep the one-per-session rule and the scratch item, and name the choices: load acta:brainstorm; offer a background agent (`claude --bg 'brainstorm SCRATCH-n'`) and a new session where the user types `brainstorm SCRATCH-n`; a herdr tab only when `HERDR_ENV=1`.
2. `plugin/hooks/default-rules.md` is generated from `coreRules`. Regenerate it with `go test ./internal/hook -run TestDefaultRulesFile -update`.
3. The eval grader and prompt do not change.

## Testing

- A red test in `internal/hook/hook_test.go`: the session-start text holds `claude --bg 'brainstorm SCRATCH-n'`, the new-session choice and `HERDR_ENV=1`.
- `go test ./...` green.
- `scripts/eval --case second-brainstorm-choices` passes 3 runs in a row.
