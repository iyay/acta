---
parent: bugs/2026-11-10-setup-wizard-wipes-unasked-keys
id: SPC-0199
created: "2026-11-10 19:10:06"
hash: gz3mb5b
---
# Setup keeps the keys it does not ask about

Bounded. The user approved this design on WhatsApp on 2026-11-10.

## Problem

`AnswersUser` at `internal/setup/form.go:286` builds a new `config.User` from the form answers only and never reads `current`. `internal/setup/run.go` saves that, so `acta setup` wipes every key the form does not ask: `theme`, `coding_guide`, `update_check`, `parallel_plans`, `subagent_picks` and `subagent_offer`. A probe in a temp worktree showed all six empty after `AnswersUser`; starting from `current` kept all six, and the `internal/setup` tests stayed green.

## Goal

After `acta setup`, every key the form does not ask about keeps the value it had when the wizard started, and every key the form asks takes the user's answer.

Out of scope, ruled by the user after the cold read:
- A config file that `LoadUser` rejects. Setup starts from defaults then, as today, and saves defaults plus the answers.
- A change made to the file while the wizard is open, such as a subagent pick from another session. Setup keeps the snapshot it read at the start.

## Design

1. `AnswersUser` starts from `current` and always sets the ten fields the form asks, even to an empty answer, so a user can still clear the tone: chat language, style, tone, build executor, plan depth, plan split, review depth, commit history, questions and review understanding. Every other field keeps the value from `current`, so a key added to `config.User` later is kept without a change here.
2. Files: `internal/setup/form.go` and `internal/setup/form_test.go`.

## Testing

- `internal/setup/form_test.go`: the test walks every field of `config.User` by reflection. With a `current` where every field is set, each field outside the ten asked ones comes back unchanged, and each asked field equals the form answer.
- One case clears the tone: `current.Tone` is `x`, the answer is empty, and the result is empty.
- At land: the full suite.
