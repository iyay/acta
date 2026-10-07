---
parent: specs/2026-10-08-omp-eval-js-regex
depth: minimal
id: PLN-0115
created: "2026-10-08 05:13:28"
hash: lcjmthh
started: "2026-10-08 05:15:35"
finished: "2026-10-08 05:16:45"
---
# omp eval runner reads grader patterns as JavaScript regex

**Goal:** The omp eval runner compiles every grader pattern as a JavaScript regex, so the probe-round grader with lookaround grades instead of failing to compile.

**Spec:** .acta/specs/2026-10-08-omp-eval-js-regex.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Engine: `github.com/dlclark/regexp2` with `regexp2.ECMAScript`. It is already in `go.mod` as indirect; `go mod tidy` moves it to the direct block. No other new dependency.
- Match timeout: 2 seconds on every compiled pattern (`MatchTimeout`).
- Do not edit any grader file under `plugin/evals`.

## Waves

- Wave 1: Task 01, Task 02

### Task 01: One JavaScript regex helper for both pattern graders

**Files:**
- Create: `internal/evalomp/pattern.go`
- Modify: `internal/evalomp/grade.go` (`gradeRegex`, `gradeTool`)
- Modify: `go.mod`, `go.sum`
- Modify: `internal/plugincheck/evals_test.go` (new test `TestEvalGraderPatternsCompileInOmp`)
- Test: `internal/evalomp/pattern_test.go`, `internal/evalomp/grade_test.go`

**verify:** No grader pattern in `internal/evalomp` is compiled or matched by Go's `regexp` any more, no match error (bad pattern, timeout) can turn into a PASS, and every `regex` and `tool_used` grader pattern under `plugin/evals` is checked by `go test`. List every place a pattern is compiled or matched and what each does on error, and the grader types the plugincheck test covers.

**Interfaces:**
- Produces: `func CompilePattern(pattern, flags string) (*regexp2.Regexp, error)`, ECMAScript mode, `IgnoreCase` when `flags` holds `i`, `MatchTimeout` 2 s. `func PatternProblems(cases []Case) []string`, one line per `regex` grader `pattern` or `tool_used` grader `input_match` that does not compile, naming case and grader.

- [x] Failing test: the real pattern from `plugin/evals/probe-round/graders/every-question-has-a-recommendation.md` (copied into the test) with `match: not_contains` FAILs on a reply where one `**Q1.** ...` block has no `Recommended:` and PASSes when every block has one; `flags: i` still matches across case; a `tool_used` `input_match` still matches as before; `PatternProblems` names a grader whose pattern is `(` and returns nothing for good ones. Also add `TestEvalGraderPatternsCompileInOmp` in `internal/plugincheck/evals_test.go`: it loads `filepath.Join(pluginRoot(t), "evals")` with `evalomp.LoadCases` and reports each line of `evalomp.PatternProblems` with `t.Error`, with a one-line comment saying why (the omp runner reads these patterns as JavaScript regex). Write `PatternProblems` first on the current Go `regexp` engine and run `go test ./internal/evalomp/ ./internal/plugincheck/ -run 'TestGrade|TestPattern|TestEvalGraderPatternsCompileInOmp'`: it fails on probe-round with `invalid or unsupported Perl syntax`.
- [x] Code: add `CompilePattern` in `pattern.go` on regexp2 and make `PatternProblems` use it; make `gradeRegex` and `gradeTool` call `CompilePattern` and treat a `MatchString` error as FAIL with the pattern in the message; drop the `(?i)` prefix and the `regexp` import; run `go mod tidy`.
- [x] Run `go test ./internal/evalomp/ ./internal/plugincheck/` passes, `go vet ./internal/evalomp/ ./internal/plugincheck/` and `gofmt -l internal` are clean, then commit.

### Task 02: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.36`, so `0.1.37`. List each file and the version it holds.

- [x] Failing test: run `go test ./internal/plugincheck/` after changing only `plugin.json` to `0.1.37`; the version-agreement check fails.
- [x] Code: set `0.1.37` in the other two files.
- [x] Run `go test ./internal/plugincheck/` passes, then commit.

## Review notes

- Round 1 CLEAN on both axes (Spec and Standards), no BLOCKER.
- `\b` in regexp2 ECMAScript mode counts non-ASCII letters as word characters while `\w` stays ASCII, so `\bfunc \w+\(` misses `éfunc f(`; only hits text with an accented letter right next to a keyword.
- Other engine differences found by a side-by-side check (`\s` matches U+00A0, `$` without `m` does not match before a trailing newline, `ſ` no longer folds to `s`) all match JavaScript; no `tool_used` `input_match` changed behaviour.
- The spec's "fixture grader" check is an in-memory case in `TestPatternProblems`, and the real plugincheck test was proved red by mutation.
- `go.sum` needed no change; it already held both regexp2 lines.
- The `m` flag drop that both reviewers found is older than this plan; it is filed as its own bug, `.acta/bugs/2026-10-08-omp-eval-drops-multiline-flag.md`.
