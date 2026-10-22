---
id: SPC-0171
created: "2026-10-22 18:44:49"
hash: oyr063d
started: "2026-10-22 18:47:07"
finished: "2026-10-22 18:47:40"
---
# Stop citing rules and skill paths in replies

## Problem

Codex adds a line like this to its reply in many rounds:

`Gate dari acta:slice (/Users/.../skills/slice/SKILL.md): "Approval of the design did not approve the plan."`

No skill asks for it. The model quotes the rule and the skill file path to
justify a step it is about to take. The user gets nothing from it, and it
shows up again and again. The acta output style has no line against it.

## Design

Bounded change approved in chat on 2026-10-22.

- `plugin/output-styles/acta.md`: add one line under `## Every reply`:
  `- Never cite a rule, skill or file path to justify a step unless asked why.`
  Claude Code reads this file as the output style. Codex gets the same text in
  the session note through `plugin/style.go`, so both harnesses get the rule.
- The style body cap in `internal/plugincheck/style_test.go` is 300 words. The
  body is 283 words today, so the 15-word line fits (298) and the cap stays.
- `internal/plugincheck/budget_test.go`: raise the byte cap of
  `output-styles/acta.md` to the new file size in the same commit.

## Testing

A new test in `internal/plugincheck/style_test.go` fails when the body has no
such line. It is written first and seen failing.

No eval case: the Go test pins the text, and an eval run costs quota for one
line.

## Out of scope

Rewording other style lines.
