---
id: BUG-0035
hash: qfgjosb
priority: low
---
# omp evals ignore the m flag, so five-questions-max can pass a reply with six questions

## Symptom
Under `scripts/eval --omp`, the probe-round grader `five-questions-max` (`flags: im`, pattern `(?:^\*\*Q6\.|^ {0,3}6\.\s)`, `match: not_contains`) passes a reply whose `**Q6.**` heading sits partway down. `^` only matches at the very start of the reply, because the `m` flag is dropped. `claude plugin eval` reads it as a JavaScript multiline regex and fails the same reply.

## Root cause
The omp runner maps only the `i` flag. On main that is `internal/evalomp/grade.go:102-104` (adds `(?i)` and nothing else); after PLN-0115 it is `CompilePattern` in `internal/evalomp/pattern.go`, which sets `IgnoreCase` for `i` and never `Multiline` for `m`.

## Repro
`CompilePattern("^\\*\\*Q6\\.", "im")` then `MatchString("intro\n**Q6.** x")` returns false; JavaScript `/^\*\*Q6\./im.test("intro\n**Q6.** x")` returns true.

## Found in
Branch omp-eval-js-regex, review round 1 of PLN-0115 (Spec and Standards reviewers). Present on main before that plan.
