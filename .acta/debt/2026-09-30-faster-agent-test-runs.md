---
id: DBT-0034
hash: iki2l1r
parent: plans/2026-09-30-faster-agent-test-runs
---
# Faster test runs for agents review notes

- [ ] (medium) scripts/test:18 `--full` runs `go run ./cmd/acta run-one`; SIGTERM to `go run` (an agent harness timeout) kills only `go run` and leaves acta and the suite running as orphans that still hold the machine lock. Fix idea: build acta first and exec the binary.
- [ ] (low) internal/cli/runone.go:43 the signal-forwarding goroutine ranges over a channel that is never closed, so each in-process call (tests) leaks one goroutine.
- [ ] (low) internal/cli/runone.go:45 a terminal Ctrl-C reaches the command twice (from the tty process group and forwarded by acta); some tools treat a second SIGINT as a hard kill.
- [ ] (low) internal/cli/runone.go:50 a command killed by a signal comes back as exit 3 with "signal: interrupt", not 128+n; the plan line "exit code passes back unchanged" does not name this case.
- [ ] (low) scripts/test:15 a two-word flag value starting with ./ (for example -coverprofile ./c.out) counts as a package, so ./... is not added; a full import path gets ./... added and tests the whole repo.
- [ ] (low) internal/write/runone.go the blocking Flock(LOCK_EX) does not retry on EINTR; Go's SA_RESTART covers it today, no failure reproduced.
- [ ] (low) internal/cli/cli.go the unknown-command list leaves out voice and hook (already so before this branch).
- [ ] (low) plugin/skills/tdd/SKILL.md:271,287 still show a bare `$ npm test` as sample RED/GREEN output; reads as an example, could be narrowed to match the narrow-tests rule.
- [ ] (low) cmd/acta/runone_test.go sets env only through cmd.Env, so it could run in parallel; the plan keeps every runone_test.go serial on purpose.
- [ ] (low) The 40s full-suite target was missed on a busy machine (45.8s and 55.4s at load 47-58, before 90.7s and 128.8s); it was never measured on a quiet machine. CPU sys stays about 107s (git process spawns), which this plan did not touch.
