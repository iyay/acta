---
parent: specs/2026-10-08-omp-eval-multiline-flag
closes: [DBT-0097.01]
depth: minimal
id: PLN-0116
created: "2026-10-08 05:25:34"
hash: oe5wgvt
started: "2026-10-08 05:26:46"
finished: "2026-10-08 05:27:42"
---
# omp eval runner honours the m flag and refuses unknown flags

**Goal:** `CompilePattern` maps `m` to multiline, refuses any flag it does not know, and the timeout test stops waiting out real 2 s timeouts.

**Spec:** .acta/specs/2026-10-08-omp-eval-multiline-flag.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Known flags are exactly `i` (`regexp2.IgnoreCase`) and `m` (`regexp2.Multiline`). Any other character is an error that names it.
- The match timeout in real runs stays 2 seconds.
- Do not edit any grader file under `plugin/evals`.

## Waves

- Wave 1: Task 01, Task 02

### Task 01: m flag, unknown flags refused, short timeout in tests

**Files:**
- Modify: `internal/evalomp/pattern.go` (`CompilePattern`, `matchTimeout`)
- Test: `internal/evalomp/pattern_test.go`, `internal/evalomp/grade_test.go` (`TestGradeMatchErrorNeverPasses`)

**verify:** No flag in a grader's `flags` is ever ignored: each one is either applied (`i`, `m`) or makes `CompilePattern` return an error naming it, and that error reaches a FAIL on every grading path. Real runs still use a 2 s timeout, and no test in `internal/evalomp` waits out a real 2 s timeout. List every flag character handled and every path the error takes.

- [x] Failing test: in `pattern_test.go`, `CompilePattern("^\\*\\*Q6\\.", "im")` matches `"intro\n**Q6.** x"` and with flags `"i"` does not; `CompilePattern("a", "s")` returns an error whose text contains `s`. In `TestGradeMatchErrorNeverPasses`, set the timeout to a short value (for example 50 ms) and restore it with `t.Cleanup`. Run `scripts/test ./internal/evalomp/`; it fails (no multiline, no error, and the timeout is a constant).
- [x] Code: turn `matchTimeout` into a package `var` with its comment kept; in `CompilePattern`, loop over the flag characters: `i` adds `IgnoreCase`, `m` adds `Multiline`, anything else returns `fmt.Errorf` naming the flag and saying only `i` and `m` are supported.
- [x] Run `scripts/test ./internal/evalomp/ ./internal/plugincheck/` passes and the evalomp time drops by seconds, `go vet ./internal/evalomp/` and `gofmt -l internal` are clean, then commit.

### Task 02: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.37`, so `0.1.38`. List each file and the version it holds.

- [x] Failing test: change only `plugin.json` to `0.1.38`; `scripts/test ./internal/plugincheck/ -run TestManifests` fails.
- [x] Code: set `0.1.38` in the other two files.
- [x] Run `scripts/test ./internal/plugincheck/ -run TestManifests` passes, then commit.

## Polish

### Task 03: Review polish

**verify:** The NOTE below is applied, and nothing else changes.

- [x] `internal/evalomp/grade_test.go`: the comment in `TestGradeMatchErrorNeverPasses` that says the pattern hits the 2 second timeout now says it hits the short test timeout set above.
- [x] Commit: `polish: review notes for PLN-0116`

## Review notes

- Round 1 CLEAN on both axes, no BLOCKER; polish review CLEAN on both axes.
- Putting the test timeout back to 2 s turns no test red; only the run time shows it (about 4 s against 0.4 s).
- With `m`, `five-questions-max` under omp now fails any reply with a line starting `6. `, a six-item numbered list included; this matches how the Claude runner reads the grader.
- No `t.Parallel` in `internal/evalomp`, so the package `matchTimeout` var has no race; `-race` passed.
- Every grader `flags` value under `plugin/evals` and `plugin/evals-routing` is `i` or `im`, so none errors under the new rule.
