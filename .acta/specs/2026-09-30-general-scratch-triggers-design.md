---
created: "2026-09-30"
parent: scratch/2026-09-29-general-scratch-triggers
id: SPC-0041
hash: mmd8j6b
started: "2026-09-30"
finished: "2026-09-30"
---
# General scratch trigger words

Status: design approved by the user in chat on 2026-09-30. Bounded: the scratch skill, the hook skills index and their tests exist already. Only the example words change.

## Why

The plugin ships to every user. The scratch trigger examples are the user's own Indonesian slang ("catet", "nanti", "kepikiran"). Other users do not talk like that. The examples must be general, and the agent must know the same intent in any language.

## Changes

1. `plugin/skills/scratch/SKILL.md`
   - Description (line 3): the list becomes `(note this, later, idea for later, or the same intent in any language)`.
   - "When to file" (line 12): the list becomes `"note this", "later", "idea for later", or the same intent in any language`.
   - "note this" must stay. The `note-to-scratch` eval guards on that phrase.
2. `internal/hook/hook.go` line 30 and `plugin/hooks/default-rules.md` line 10: the scratch line becomes `raw ideas ("note this", "later", side ideas, any language); file with acta scratch new, never memory`. Both texts stay the same string.

## Not changed

Test fixtures that use "catet" as body text stay: `internal/write/ops_test.go`, `internal/write/scratch_test.go`, `internal/board/testdata/basic/.acta/scratch/`, `internal/tui/sidebar_test.go`. They test bytes and layout, not trigger words.

## Testing

- `internal/plugincheck/skill_scratch_test.go`: `Must` drops "catet", "nanti", "kepikiran" and adds "note this", "idea for later", "any language". `MustNot` adds "catet", "nanti", "kepikiran" so the slang cannot come back.
- `internal/hook/hook_test.go` lines 36 and 140: expected text becomes the new scratch line.
- Both tests go red first against the old text, then green.
- `plugin/evals/note-to-scratch` still finds "note this" (`internal/plugincheck/evals_test.go`).
