---
parent: bugs/2026-10-08-omp-eval-probe-round-lookahead
id: SPC-0106
created: "2026-10-08 05:11:54"
hash: zd4p2ir
started: "2026-10-08 05:15:35"
finished: "2026-10-08 05:16:45"
---
# The omp eval runner reads grader patterns as JavaScript regex

Status: Bounded, approved by the user in chat on 2026-10-08. Fixes BUG-0034.

Why: the suite says a `regex` grader holds a JavaScript regex (`plugin/evals/FACTS.md`, grader table). The omp runner compiles grader patterns with Go's `regexp` (RE2) in `internal/evalomp/grade.go` (`gradeRegex`, `gradeTool`). RE2 has no lookaround, so `plugin/evals/probe-round/graders/every-question-has-a-recommendation.md` fails with `invalid or unsupported Perl syntax: '(?!'` on every omp run, whatever the skill says.

Design:
- One compile helper in `internal/evalomp` uses `github.com/dlclark/regexp2` with the `ECMAScript` option. It is already in `go.mod` as an indirect dependency; it becomes direct.
- `gradeRegex` (`pattern`) and `gradeTool` (`input_match`) both use that helper, so every grader pattern goes through the same engine.
- `flags: i` maps to the `IgnoreCase` option instead of a `(?i)` prefix.
- Each compiled pattern gets a `MatchTimeout` of 2 seconds, since regexp2 backtracks. A timeout or other match error makes the grader FAIL with a message that names the pattern.
- `internal/plugincheck` gains a check: every `regex` grader under `plugin/evals` compiles with the omp helper, so a broken pattern fails `go test` before any eval runs.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: rewriting any grader pattern; changing the Claude eval path.

Tests: the real probe-round grader pattern compiles; it FAILs on a reply where one `**Qn.**` block has no `Recommended:` and PASSes when every block has one; `flags: i` still matches across case; a `tool` grader `input_match` still matches as before; the plugincheck check fails on a fixture grader with a pattern that does not compile.
