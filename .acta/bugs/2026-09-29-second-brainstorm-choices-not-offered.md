---
id: BUG-0006
hash: obj8k72
started: "2026-09-29"
fixed_in: 231afdc
finished: "2026-09-29"
---
# Second big brainstorm gets made-up options instead of the two real choices

## Symptom
When a session already has an open Architectural brainstorm and the user asks for a second big change, the agent files nothing concrete and offers made-up options ("finish billing first", "park billing", "leave it as scratch"). It does not offer the background agent (`claude --bg 'brainstorm SCRATCH-n'`) or the manual new session. Eval case `second-brainstorm-choices` fails most runs (2/2 on main at 70f89de, 2/3 on the PLAN-30 branch).

## Root cause
The agent never loads `acta:brainstorm`. The trace of a failing run has zero tool calls: it answers from the session-start rule alone. That rule (`plugin/hooks/default-rules.md:23`, same text in `internal/hook/hook.go:53`) says only "the user picks how to open it" and names no choices, so the agent invents them. The choices live only in `plugin/skills/brainstorm/SKILL.md` ("Architectural path: scratch item and one per session").

## Repro
`scripts/eval --case second-brainstorm-choices --keep-temp`, then read `<kept>/out/trace.jsonl`: no tool_use entries; the reply lists invented options.

## Found in
main, while landing PLAN-30 (eval gate), then acta:debug on 2026-09-29.
