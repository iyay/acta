---
id: SCR-0034
hash: xki1yk5
title: Stop parallel and orphaned test runs from pinning the CPU
status: brainstorming
created: "2026-10-01 15:41:57"
schema: "1"
started: "2026-10-01 15:42:04"
finished: "2026-10-01 15:50:25"
---
# Stop parallel and orphaned test runs from pinning the CPU

## Words

### 2026-10-01

Fix for bugs/2026-10-01-parallel-test-runs-pin-cpu.

User asked (2026-10-01) to brainstorm three fixes:

1. scripts/test: the short path also goes through the acta run-one lock.
2. Agents must not run bare `go test`; use scripts/test. Maybe the pre-tool
   hook blocks a bare `go test ./...` that skips run-one.
3. A test binary stops by itself when its `go` parent dies, so a killed run
   leaves no orphan.

## Context

## Log

### 2026-10-01

Q1 lock scope: user picked (a). Only whole-repo runs (scripts/test with no package, ./...) take the run-one lock. Narrow runs like scripts/test ./internal/write -run X stay free so the TDD loop never waits.

### 2026-10-01

Q2 hook scope: user picked (b). In a repo that has scripts/test, pre-tool blocks every bare go test (any package, any flags) and the block message names the scripts/test command to use instead. Only real command starts match (start, after && ; |, after a prefix like rtk), so text such as grep "go test" is not blocked.

### 2026-10-01

Q3 orphans: user picked (a). A small shared helper called from TestMain in every package with tests. A goroutine checks os.Getppid() every second and exits the binary when the parent changes. One test checks that every test package calls the helper.

### 2026-10-01

Section 1 approved (scripts/test): whole-repo runs (no package, or ./...) run as go run ./cmd/acta run-one -- go test -short ./... under the same lock as --full; package runs stay exec go test -short with no lock; --full unchanged; the waiting line still prints. Tests in scripts/ cover the three cases.

### 2026-10-01

Section 2 approved (pre-tool block): active when the git top-level of the hook cwd has scripts/test; blocks go test at a real command start (start, after && || ; | (, after prefixes rtk, env X=1, time); lets through plain text, scripts/test and acta run-one -- go test; exit 2 with an English message that names the scripts/test command, carrying the package and -run over; works for Claude Code and omp through the same exit-2 path; one line in build/SKILL.md and the dispatch.md brief. Table test in internal/hook.

### 2026-10-01

Section 3 approved (orphans): new internal/testguard with Watch(): saves ppid at start, goroutine checks every second, on change prints 'testguard: parent process is gone, stopping' and exits 1. Added to all 15 test packages (11 new TestMain, 4 existing get a first line). A plugincheck test fails when a folder with *_test.go does not call testguard.Watch(). Watch's own test orphans a child through sh -c ... & and waits up to 5 s for it to go.

### 2026-10-01

Section 4 approved: hook fails open on any error; rollout needs land + go install ./cmd/acta (say it in the land report), skill text after restart; eval sandbox has no scripts/test so it is not blocked, scripts/eval still runs at land; out of scope: other languages, -p or GOMAXPROCS limits, killing current orphans, changing run-one.

## Open questions
