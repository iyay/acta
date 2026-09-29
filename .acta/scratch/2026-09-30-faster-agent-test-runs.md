---
id: SCR-0029
hash: fbcrlqa
title: 'Faster test runs for agents: narrow verify runs and one full suite at a time'
status: brainstorming
created: "2026-09-30"
schema: "1"
started: "2026-09-30"
finished: "2026-09-30"
---
# Faster test runs for agents: narrow verify runs and one full suite at a time

## Words

### 2026-09-30

## Words

User (2026-09-30): "check SPC-0025 deh, ini kok tests masih lama ya. apalagi kalo agent yang running, lama banget" ... "masa buat running full suite bisa sampe 5 menit sendiri" ... "brainstorm B+C".

B = verify runs during build use `go test -short ./...` or per-package tests; the full suite runs only at the acta:land gate.
C = a machine-wide lock so only one full suite runs at a time across worktrees.

## Context

Measured 2026-09-30 on 8 cores, 16 GB: `go test -count=1 -short ./...` 67s, full 100s alone.
At 06:15 three agent worktrees (tui-colors, batch-author-lookup, config-file-name) ran `go test ./...` at once. Load average 26, about 128 MB RAM free, swapping, Docker VM busy. One suite stretched to 3-5 minutes.
Plans carry `go test ./...` in verify lines (62 times in 2026-09-29/30 plans). plugin/skills/build/SKILL.md:116 says `go test ./...`.

## Context

## Log

### 2026-09-30

- 2026-09-30 context: a plan run triggers the full suite many times (build baseline, per-task verify lines, end of build, land precondition, post-merge). flock lock already exists at internal/write/tick.go:202. Q1 asked: C in acta CLI (acta run-one -- cmd) or repo-only scripts/test.

### 2026-09-30

- 2026-09-30 user answer Q1: both. acta CLI gets the lock command (works in every repo), and this repo gets scripts/test that uses it.

### 2026-09-30

- 2026-09-30 user answer Q2: (a) full suite only in acta:land (precondition, plus post-merge only when the merge tree differs from the branch tree); everything else -short or touched package. User adds: the full suite itself is still too long.
- 2026-09-30 measured: internal/tui full alone 57.6s wall, user 67s sys 44s, so about 2 of 8 cores busy. Top: TestViewNeverOverflowsAnyWindow 40.7s, TestNoRoundedCorners 24.5s, TestThumbSitsOnTheBorderNotInside 14.1s, each one serial loop. Git packages serial because helpers call t.Setenv (write 32s, cli 23s, cmd 25s summed).

### 2026-09-30

- 2026-09-30 user answer Q3: (a) parallelize without changing what tests check: split sweep loops into parallel subtests, drop t.Setenv from git helpers (repo-local git config). User asked why, since tests already use t.Parallel: top-level parallel only; one 41s serial loop bounds tui wall and leaves 7 cores idle; git packages mostly serial because t.Setenv forbids t.Parallel.
- 2026-09-30 design section 1 presented: acta run-one -- <cmd>, one flock per user per machine in <cache>/acta/locks/run-one.lock reusing lockDir/safeDir, waits with no timeout, one stderr line while waiting, passes stdio and exit code. scripts/test: -short by default, --full wraps go run ./cmd/acta run-one -- go test ./..., other args go to go test.

### 2026-09-30

- 2026-09-30 user approved design section 1 (lock + scripts/test).

### 2026-09-30

- 2026-09-30 design section 2 approved: narrow tests during build (touched package, or fast mode), full only in acta:land; scripts/test when present; plan verify lines narrow; build baseline + end of build fast; dispatch GATES same; land precondition full under acta run-one, post-merge rerun skipped when merge tree equals branch tree; reviewers run only tests that prove a finding. In-flight plans not rewritten. Eval gate will run at land.

### 2026-09-30

- 2026-09-30 design section 3 approved: split 5 sweep tests into parallel per-width subtests (fresh newModel per size), git identity by repo-local git config instead of t.Setenv, PM_ROOT once in TestMain, t.Parallel where no Setenv/Chdir left, verify full + race short, timings under run-one, target full alone 40s or report real numbers, fix watch_test.go:60 race if -race trips.

## Open questions
