---
parent: specs/2026-10-28-offline-session-usage-design
depth: minimal
id: PLN-0199
created: "2026-10-28 23:11:46"
hash: qtu2gre
started: "2026-10-28 23:16:28"
---
# Offline Codex session usage implementation plan

**Goal:** Add an offline `acta usage [--json]` command that counts recorded Codex usage by repository, root session, and model without counting responses twice.
**Spec:** `.acta/specs/2026-10-28-offline-session-usage-design.md`
**Tests:** Fast `scripts/test`; full `scripts/test --full`; task loops use only the focused commands below.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Implement inline in a worktree after plan approval, one commit per task; never push or change the pre-release version for this plan.
- Follow SPC-0182's exact input field paths, JSON version 1 contract, counters, ordering, UTC ranges, partial/no-data semantics, and privacy rules; synthetic fixtures only, with no real conversation content in the repo.
- Use standard-library Go and existing Git, help/man, JSON-output, and testguard facilities; no new dependency, generic adapter framework, runtime hook, database, pricing, TUI screen, or optimization mechanism.
- Read existing transcripts without network access, background update probes, or writes to source files or user state; root sessions include attributed child usage, and reasoning is never added to output again.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2, after Task 1.
- Wave 3: Task 3, after Task 2.

### Task 1: Read supported Codex transcript events

**Files:** Create `internal/usage/log.go`, `internal/usage/log_test.go`, and `internal/usage/main_test.go`.
**Seam:** `usage.ReadLog(name string, r io.Reader) (usage.Log, error)` returns the first usable file identity and starting directory, owned response records with model/timestamp/usage validity, and located diagnostics; it performs no filesystem or Git discovery.
**verify:** At the stream boundary, only supported nested fields establish identity, model, timestamp and per-response usage; later metadata cannot change ownership, inherited records and cumulative counters never become owned usage, missing/empty/non-string model context resets to unknown, RFC 3339 timestamps normalize to UTC while invalid timestamps preserve tokens as undated, required counters reject missing/negative/fractional/out-of-int64 values and cached input above input, optional reasoning/cache-write fields add no tokens, unknown event types are ignored, complete malformed lines are recoverable, an unterminated final line stays pending, and reader failures return an error without retaining conversation text; list the adversarial paths checked.
- [x] Red: Add `TestReadLog` fixtures and the required testguard `TestMain`, then watch `scripts/test ./internal/usage -run '^TestReadLog'` fail because the supported reader is absent.
- [x] Change: Implement the streaming reader and its data types in `log.go`, preserving response IDs only in internal records for later deduplication, then run `scripts/test ./internal/usage -run '^TestReadLog'` to green.
- [x] Commit: `feat(usage): read Codex response usage events`

### Task 2: Build repository-scoped session usage reports

**Files:** Create `internal/usage/usage.go`, `internal/usage/usage_test.go`, `internal/usage/report.go`, and `internal/usage/report_test.go`; consume Task 1's `Log` without modifying its parser.
**Seam:** `usage.Read(cwd, sourceDirectory string) (usage.Report, error)` discovers and snapshots logs, calls `ReadLog`, resolves repository membership, and produces the complete JSON-tagged report; expose `usage.ErrNotRepository` so CLI wiring can distinguish bad input from operational failures, and keep partial results in `Report.Coverage.Partial` rather than returning a fatal error.
**verify:** At the filesystem/Git boundary, the current checkout and registered existing worktrees share scope, nested/unrelated repositories stay separate, child records follow established root membership while missing-root groups require their own matching directory and remain partial, and paths containing spaces or shell syntax remain literal arguments; scanning skips symlinks and archived/compressed logs, respects captured file sizes, distinguishes missing sources from unreadable sources/files and unclassified metadata, and writes nothing; report aggregation counts each response once, excludes all conflicting copies regardless of scan order, separates models and root groups, uses checked int64 sums and fresh input subtraction, preserves unknown models, produces exact coverage/status/range/order/diagnostic semantics and the canonical partial/no-data JSON examples, and represents unavailable usage without claiming measured zero; list the membership, conflict, snapshot and reporting paths checked.
- [x] Red: Add `TestUsageRead` synthetic temporary repositories/log trees and `TestUsageReport` contract checks, then watch `scripts/test ./internal/usage -run '^TestUsage(Read|Report)'` fail because repository reporting is absent.
- [x] Change: Implement scoped discovery and aggregation in `usage.go` using existing `gitc.Worktrees` and `gitc.CommonDir` where they fit, define report types and stable JSON fields in `report.go`, and run `scripts/test ./internal/usage -run '^TestUsage(Read|Report)'` to green plus `scripts/test ./internal/usage` for parser integration.
- [x] Commit: `feat(usage): aggregate repository session usage`

### Task 3: Deliver the offline usage CLI

**Files:** Create `internal/cli/usage.go` and `internal/cli/usage_test.go`; modify `internal/cli/cli.go`, `internal/cli/help.go`, `internal/cli/help_test.go`, `internal/cli/notice.go`, and `internal/cli/notice_test.go`.
**Seam:** `cli.Run([]string{"usage", ...}, stdin, stdinIsTTY, stdout, stderr)` through new `cmdUsage(args []string, stdout, stderr io.Writer) int`, consuming `usage.Read` and the existing `printJSON`, help table, exit codes, and notice-test hooks.
**verify:** CLI wiring proves one supported success and one fatal error passed through from the reader without repeating lower-layer fixture cases; command-owned behavior resolves empty/unset/relative CODEX_HOME correctly, rejects positionals/unknown flags and outside-Git execution, maps no-data/complete/partial/operational results to exits 0/0/2/3 and bad input to 1, renders the specified human tables and no-data text with control characters escaped, preserves JSON as a single report with structured diagnostics and no transcript/response-ID disclosure, handles output-writer failure, publishes scope limitations through help/man, and starts no update probe or user-cache write on valid or invalid usage invocations; inspect all command entry paths while preserving ordinary commands and TUI startup behavior.
- [ ] Red: Add `TestUsageCommand` and `TestUsageRender` command-owned checks plus `TestUsageNoUpdate` in the notice tests, then watch `scripts/test ./internal/cli -run '^TestUsage(Command|Render|NoUpdate)'` fail because usage is not routed and its offline contract is absent.
- [ ] Change: Wire and render `usage`, add its help-table entry, exempt usage from automatic update checks/notices, and run `scripts/test ./internal/cli -run '^TestUsage|^TestHelp|^TestEveryRunCase|^TestUpdateCheck|^TestUpdateNotice|^TestTUIBoardsFirstFrameIsMainOnly'` to green.
- [ ] Commit: `feat(cli): add offline session usage command`

## Verification evidence

During build, record each task's observed red and green results here. Before landing, compare the built command with an independent offline sum over the available historical spike logs, separating the active session and other newly added logs rather than pinning mutable historical totals in tests. Compare the same set of closed historical files in both calculations so new activity cannot explain a difference. Record the compared scope, models, response counts, fresh/cached/output totals, and any coverage limitation. If those historical files are gone, say so and retain the synthetic fixture evidence. Full landing gates run after implementation review; this plan authorizes no paid agent calls.
