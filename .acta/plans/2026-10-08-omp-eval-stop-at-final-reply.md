---
parent: specs/2026-10-08-omp-eval-stop-at-final-reply
depth: minimal
id: PLN-0117
created: "2026-10-08 06:52:20"
hash: gg87x09
started: "2026-10-08 06:53:46"
finished: "2026-10-08 06:57:32"
---
# omp eval runner stops at the first final reply and runs omp in a clean home

**Goal:** A finished omp eval case ends at its first final reply instead of running into the timeout, and omp no longer sees the user's real home apart from `~/.omp`.

**Spec:** .acta/specs/2026-10-08-omp-eval-stop-at-final-reply.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- No test may run the real omp or touch the real home: tests use `fakeOmp` in `internal/evalomp/run_test.go` and point `HOME` at a temp dir.
- Do not change `OmpJudge`, any case, prompt or grader.

## Waves

- Wave 1: Task 01, Task 02

### Task 01: Stop at the first final reply, run omp in a clean home

**Files:**
- Modify: `internal/evalomp/run.go` (`RunCase`), `internal/evalomp/result.go` (`ParseStream`, `streamEvent`)
- Test: `internal/evalomp/run_test.go`, `internal/evalomp/result_test.go`

**verify:** A run ends at the first assistant `message_end` with `stopReason: stop`, or at `agent_end` if that comes first, and never counts a reply or tool call after that point; omp's whole process group is gone when `RunCase` returns on every path (final reply, `agent_end`, timeout, error); omp sees `HOME` as the throwaway home, where only `.omp` links to the real `~/.omp`. List every way a run can end and what `RunCase` returns and kills on each.

- [x] Failing test: a fake omp that prints one tool call, a final `message_end` (`stopReason: stop`, text "done"), one more tool call, then `sleep 30`, run with `TimeoutSeconds: 5`: `RunCase` returns reply "done", one call, no error, in well under 5 s, and the sleep child is gone (copy the child check from `TestRunCaseTimeoutKillsChildren`). A fake omp that prints `$HOME` and `readlink "$HOME/.omp"` into its log dir, with the test's `HOME` set to a temp dir holding `.omp`, records the throwaway home and a link to that temp `.omp`. A stream with neither a final reply nor `agent_end` still errors (keep the existing case). Run `scripts/test ./internal/evalomp/`; the new tests fail today.
- [x] Code: `ParseStream` stops reading at the first assistant `message_end` with `stopReason` `stop` (reply = its text parts) or at `agent_end`; `RunCase` feeds omp's stdout through a pipe into it while omp runs, then kills the process group once it returns, keeping the timeout and error paths; before omp starts, `RunCase` makes `<home>/.omp` a symlink to the real `~/.omp` (read the real home with `os.UserHomeDir` before it is changed) and sets `HOME` to `<home>` for omp.
- [x] Run `scripts/test ./internal/evalomp/` passes, `go vet ./internal/evalomp/` and `gofmt -l internal` are clean, then commit.

### Task 02: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.38`, so `0.1.39`. List each file and the version it holds.

- [x] Failing test: change only `plugin.json` to `0.1.39`; `scripts/test ./internal/plugincheck/ -run TestManifests` fails.
- [x] Code: set `0.1.39` in the other two files.
- [x] Run `scripts/test ./internal/plugincheck/ -run TestManifests` passes, then commit.

## Review notes

- Round 1 CLEAN on both axes, no BLOCKER, no [fix] NOTE; three mutations each went red; -race -count=3 clean.
- The group kill after a normal exit and reap could in theory reach a reused process group id; the odds are close to zero.
- A missing ~/.omp gives a dangling link and omp's own vague error instead of a clear one; running omp at all needs ~/.omp.
- A final reply followed by a non-zero omp exit now passes with that reply; this is what the spec asks.
- The plan's test numbers (5 s timeout) became 10 s with a 4 s check in the code; same property.
- os.RemoveAll on the case folder removes the .omp link and does not follow it into the real ~/.omp.
