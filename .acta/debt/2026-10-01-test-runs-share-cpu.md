---
id: DBT-0065
hash: l30o9g8
parent: plans/2026-10-01-test-runs-share-cpu
---
# Review NOTEs: Test Runs Share The CPU Implementation Plan

- [ ] (medium) internal/plugincheck/testguard_test.go: skipped folders return nil, not filepath.SkipDir, so .git, testdata, node_modules and plugin are still walked; a _test.go with no Watch under testdata or plugin fails the check.
- [ ] (medium) plugin/skills/plan/SKILL.md:88 still shows `go test ./internal/tui/ -run TestX` as the example run step; in repos with scripts/test the hook now blocks that shape.
- [ ] (medium) internal/hook/gotest.go: a newline is not a command separator, so `cd x` then a newline then `go test ./...` slips past.
- [ ] (low) internal/hook/gotest.go: `timeout 600`, `nice`, `bash -c '...'`, `go -C dir test` and a full path to go are not seen as go test.
- [ ] (low) internal/hook/gotest.go: command substitution inside double quotes (`echo "$(go test ./...)"`) and backticks now slip past; round 1 code blocked the first.
- [ ] (low) internal/hook/gotest.go: `$'...'` is read as a plain single-quoted string, so a `\'` inside it ends the quote early.
- [ ] (low) internal/hook/gotest.go: the Use: suggestion can name a different run: `go test -run X`, `go test . -run X` and an import path suggest a whole-repo run; `cd internal/hook && go test ./...` suggests `scripts/test ./...`.
- [ ] (low) internal/hook/gotest.go: isPackageArg treats any word with / as a package, so `go test -c -o /tmp/x.test ./internal/hook` suggests `scripts/test /tmp/x.test ./internal/hook`.
- [ ] (low) internal/hook/gotest.go: the repo check uses the hook working folder, not the command's, so `cd /other/go-repo && go test ./...` from an acta session is blocked with acta's scripts/test in the message.
- [ ] (low) internal/hook/gotest.go: the run-one check matches anywhere in the piece and is close to dead code, because the first two words already decide.
- [ ] (low) internal/hook/gotest.go: `proxy` counts as a prefix on its own, so `proxy go test` is blocked.
- [ ] (low) internal/hook/gotest.go: the comment that an escaped byte "cannot end a quote" does not match what the code does outside quotes; the split helper parameter name `at` reads badly.
- [ ] (low) internal/hook/gotest.go: words split only on space, tab and newline, where strings.Fields also split on \r, \v, \f and Unicode spaces.
- [ ] (low) internal/hook/gotest.go: chainCut in session.go was not reused; the new separator set is a wider copy.
- [ ] (low) plugin/hooks/pre-tool: a Go panic exits 2, which the hook reads as a block; the scanner cannot panic today (300k random inputs), but the exit-code mapping is fragile.
- [ ] (low) internal/cli/hook.go: the cmdHook doc comment still says the only exception is pre-tool stopping a second brainstorm.
- [ ] (low) internal/plugincheck/testguard_test.go: the check is a plain string match, so a comment that mentions testguard.Watch() satisfies it.
- [ ] (low) internal/plugincheck/testguard_test.go: the walk covers untracked folders too, so a nested worktree or scratch copy inside the checkout would fail it.
- [ ] (low) internal/testguard/testguard_test.go: running(pid) uses kill 0, which also succeeds on a zombie; under a non-reaping subreaper the 5 second check could flake.
- [ ] (low) cmd/acta/main_test.go: the testguard import sits in the stdlib import group.
- [ ] (low) scripts/test: whole, all and pkg are three flags for one decision; the header line about package runs skipping the wait is only true without ./...
- [ ] (low) scripts/test: whole-repo short runs now go through go run ./cmd/acta, so a compile error in cmd/acta or its imports hides the test output, and go run maps every failing exit code to 1.
- [ ] (low) scripts/test_test.go: no test covers exit code or quoted args passing through; checked by hand with a fake go that exits 7.
- [ ] (low) Open sibling branches that add a new test package turn main red at merge until they call testguard.Watch().
