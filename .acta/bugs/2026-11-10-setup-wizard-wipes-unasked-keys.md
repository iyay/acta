---
id: BUG-0083
hash: gr1njw3
started: "2026-11-10 19:31:23"
fixed_in: 05ae788aa7a1c8d4e362491f7deba31f48d2b6a8
finished: "2026-11-10 19:37:45"
---
# Running acta setup again wipes settings it does not ask about

## Symptom
After `acta setup` (the terminal wizard), keys the form does not ask about are gone from `~/.acta/config.yaml`: `coding_guide`, `theme` and `update_check` today, and `parallel_plans`, `subagent_picks` and `subagent_offer` once PLN-0223 lands. A user who declined the subagent default offer gets it offered again after setup, which breaks SPC-0197 ("a declined offer never comes back").

## Root cause
`internal/setup/form.go:286` `AnswersUser(current, d)` never reads `current`. It builds a new `config.User` from the form answers only, and `internal/setup/run.go:31` saves it, so every field not copied is lost.

## Repro
Set `theme: <any theme>` in `~/.acta/config.yaml` (or `subagent_picks: 3` after PLN-0223), run `acta setup` and finish the form, then read the file: the key is gone. Unit level: call `AnswersUser` with a `current` that has `Theme`, `CodingGuide`, `UpdateCheck` set; the result has them empty.

## Found in
Branch `parallel-plans-config` (PLN-0223), review round 1, both reviewers as a pre-existing item.
