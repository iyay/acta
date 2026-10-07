---
id: BUG-0034
hash: isdvbmo
fixed_in: c2886781662644139e680028c0519d40b3594457
finished: "2026-10-08 05:23:10"
---
# probe-round always fails when the evals run through omp

## Symptom
`scripts/eval --omp` reports `FAIL probe-round: every-question-has-a-recommendation (bad pattern ...: error parsing regexp: invalid or unsupported Perl syntax: '(?!')`, on main and on any branch, whatever the skill text says.

## Root cause
`plugin/evals/probe-round/graders/every-question-has-a-recommendation.md:3` uses a negative lookahead `(?!Recommended:)` and a lookahead `(?=...)`. The omp runner compiles grader patterns with Go's `regexp` at `internal/evalomp/grade.go:105-107`, and Go's RE2 syntax has no lookaround, so the grader can never pass under omp. `claude plugin eval` accepts the pattern, so the Claude path is not affected.

## Repro
`go run ./cmd/acta eval-omp --case probe-round plugin` from the repo root.

## Found in
main, while landing PLN-0114 on 2026-10-07: the omp eval was run because the Claude eval sandbox is blocked on this machine, and the same case was red on main too.
