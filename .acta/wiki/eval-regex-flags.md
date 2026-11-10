---
type: Gotcha
title: Eval regex graders take only the flags line
description: "A regex grader with an inline (?i) throws at run time; use the flags line, and [\\s\\S] in place of the s flag"
paths: [plugin/evals/]
---

A `type: regex` grader whose pattern starts with an inline group such as `(?i)` or `(?is)` passes every Go test, then throws in `claude plugin eval` with `Invalid regular expression: unrecognized character after (?`, and the case scores as a fail. Put the flags on their own `flags:` line instead.

Only `i` and `m` are allowed there: `TestEvalGraderPatternsCompileInOmp` rejects `s`, because omp cannot compile it. To match across lines, write `[\s\S]*` in place of `.*`, in a single-quoted YAML string so the backslashes stay as they are. Seen on 2026-11-10 in the `slice-split-parallel-*` cases.
