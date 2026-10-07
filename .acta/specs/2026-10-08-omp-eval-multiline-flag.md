---
parent: bugs/2026-10-08-omp-eval-drops-multiline-flag
closes: [DBT-0097.01]
id: SPC-0107
created: "2026-10-08 05:25:09"
hash: z1gqg7h
started: "2026-10-08 05:26:46"
finished: "2026-10-08 05:27:42"
---
# The omp eval runner honours the m flag and refuses flags it does not know

Status: Bounded, approved by the user in chat on 2026-10-08, with DBT-0097 folded in. Fixes BUG-0035.

Why: `CompilePattern` (`internal/evalomp/pattern.go`) reads only the `i` flag. `plugin/evals/probe-round/graders/five-questions-max.md` uses `flags: im` with `^\*\*Q6\.`, so under omp `^` matches only at the start of the reply and a reply with a sixth question further down still passes. Any other flag is dropped the same silent way. Separately, `TestGradeMatchErrorNeverPasses` waits out two real 2 s timeouts, which makes the `evalomp` package about 4 s slower (DBT-0097.01).

Design:
- `CompilePattern` maps `m` to `regexp2.Multiline`, next to `i` to `IgnoreCase`.
- Any other character in `flags` makes `CompilePattern` return an error that names the flag. Through the existing paths this FAILs the grader with the pattern named, and `TestEvalGraderPatternsCompileInOmp` fails on it, so no flag can be ignored in silence again.
- The 2 s match timeout becomes a package variable, so the timeout test can set a short one and put it back after. The value used in real runs stays 2 s.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: supporting `s`, `u`, `g` or `y`; editing any grader file.

Tests: `^\*\*Q6\.` with `im` matches `intro\n**Q6.** x` and with `i` alone does not; flags `s` returns an error naming `s`; `TestGradeMatchErrorNeverPasses` still proves a timeout FAILs, and the `evalomp` package no longer spends seconds waiting on it.
