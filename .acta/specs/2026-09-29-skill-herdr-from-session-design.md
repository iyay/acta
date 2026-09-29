---
parent: bugs/2026-09-29-brainstorm-skill-offers-herdr-tab
id: SPC-0026
created: "2026-09-29"
hash: x9gb1tp
---
# Brainstorm Skill Takes the Herdr Choice From the Session Text

Status: approved by the user on 2026-09-29 (Bounded). Fixes bug `.acta/bugs/2026-09-29-brainstorm-skill-offers-herdr-tab.md`.

## Why

- `plugin/skills/brainstorm/SKILL.md`, choice (b), says "offer this only when `HERDR_ENV=1` is set". The agent often cannot read the environment, so it offers the herdr tab "if HERDR_ENV=1" anyway.
- Eval case `second-brainstorm-choices` fails such replies. It failed about 2 of 12 runs after PLN-0033.
- PLN-0033 already moved the herdr check into the hook: the session text has a `herdr:` line only inside herdr. The skill should read that line, not the environment.

## Design

1. `plugin/skills/brainstorm/SKILL.md`, choice (b): offer a new herdr tab only when the session text has the `herdr:` line, which the acta hook adds only inside herdr. With no such line, do not mention herdr. The skill no longer names `HERDR_ENV`.
2. `internal/plugincheck/skill_brainstorm_test.go`: `HERDR_ENV=1` moves from Must to MustNot; the new wording is a Must.
3. The eval grader and prompt do not change.

## Testing

- The skill test goes red first, then green.
- `go test ./...` green.
- `env -u HERDR_ENV scripts/eval --case second-brainstorm-choices` passes 5 runs in a row.
