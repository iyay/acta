---
created: "2026-10-19 13:44:56"
id: SPC-0161
hash: doaw5p7
started: "2026-10-19 14:05:56"
finished: "2026-10-19 14:07:00"
---

# Unslop for spec self-review

Approved in chat on 2026-10-19. This is a Bounded change to the existing
`acta:shape` spec self-review.

## Behavior

Before asking the user to review a written spec, shape applies a shared
Unslop reference during its existing self-review. This covers both Bounded
and Architectural specs. The four existing checks remain. There is no
extra workflow step or second editing pass.

The reference adapts the user's Unslop rules into short, plain English.
Keep the stable rule numbers so other skills can cite them later. Treat
language patterns as reasons to inspect a sentence, not automatic errors.
Replace filler, vague claims, abstract metaphors and forced framing with
concrete facts, instructions or supported numbers. Preserve the intended
meaning and tone. Do not invent sources or measurements.

Use complete sentences when fragments would make the reader guess.
Labels and brief status text may stay short. Preserve requirements,
constraints, uncertainty, evidence, risks and approval gates. Editing prose
must not change code, commands, flags, identifiers, literal errors or direct
quotes. Existing project rules still govern required spec content.

## Files and verification

- Add `plugin/references/unslop.md` as a shared reference, not a new
  invocable skill or user setting.
- Update `plugin/skills/shape/SKILL.md` to read that reference during spec
  self-review on both paths, without changing approval gates.
- Extend `internal/plugincheck/skill_shape_test.go` to guard both paths,
  the reference link and the preservation rules. Watch the new checks fail
  before editing the plugin text.
- Register the reference in `internal/plugincheck/budget_test.go` and adjust
  only affected size limits where needed. Existing plugin embedding must
  include the reference; verify with the embed tests.
- Run focused plugin checks, embed tests and `scripts/test`. The full suite
  runs at land. Do not run paid behavior evals unless asked. Static checks
  prove the instructions ship, not that every model edit preserves meaning.

## Scope

No changes to chat style, setup, slice, build, review or land instructions.
No automatic prose fixer, new configuration or lint dependency. These can
be considered after evaluating the change in shape.

## Pilot measurement

After landing, compare the draft and edited versions of the next real spec.
Record word counts, concrete edits and whether every requirement, constraint,
source, number and approval gate kept its meaning. Fewer words alone do not
count as success. Report this as one observed sample, not evidence that every
model will behave the same way. Expanding to other steps needs that report
and a separate approval; it does not happen in this plan.
