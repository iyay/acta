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

After `acta setup`, every key the form does not ask about keeps the value it had, and every key the form asks takes the user's answer.

## Design

1. `AnswersUser` starts from `current` and sets only the fields the form asks: chat language, style, tone, build executor, plan depth, plan split, review depth, commit history, questions and review understanding. A key added to `config.User` later is kept without a change here.
2. Files: `internal/setup/form.go` and `internal/setup/form_test.go`.

## Testing

- `internal/setup/form_test.go`: a `current` with all six unasked keys set comes back from `AnswersUser` with the same six values, and each asked field holds the form answer, not the old value.
- At land: the full suite.
