---
parent: specs/2026-11-12-usage-telemetry-design
depth: minimal
id: PLN-0228
created: "2026-11-12 06:31:58"
hash: bymvb60
started: "2026-11-12 06:35:55"
finished: "2026-11-12 06:45:12"
---
# Telemetry core: counters, flush and the telemetry command

**Goal:** acta counts commands and a daily settings snapshot per local day under `~/.acta/telemetry/`, flushes finished days to the build-time endpoint with the walkie receipt rules, and gives the user `acta telemetry on|off|status`, all silent in a build with no endpoint.

**Spec:** `.acta/specs/2026-11-12-usage-telemetry-design.md` (SPC-0201)

**Tests:** fast `scripts/test ./internal/telemetry ./internal/cli`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- New package `internal/telemetry` holds every rule of the spec's "What is counted", "Consent and notice" and "Sending" sections; `internal/cli` only calls it.
- The endpoint is a package variable set by `-X`; tests swap it and point it at an `httptest` server. No test ever reaches a real network address.
- Tests set `HOME` (or the telemetry dir) to `t.TempDir()`; no test touches the real `~/.acta`.
- A telemetry failure never prints to the user, never changes an exit code, and never counts as anything.
- Reuse `internal/update.NewClient` style for the HTTP client only if it fits the Google redirect rule; otherwise a small client in `internal/telemetry`.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3
- Wave 4: Task 4

### Task 1: Telemetry store

**Files:** `internal/telemetry/store.go` (new), `internal/telemetry/store_test.go` (new)
**Seam:** `telemetry.Open(dir)` and its `Add(group, key)`, `SetSettings`, `On`, `Off`, `Enabled`, `Days` and `MarkNoticed` methods, plus the `Endpoint` package variable
**verify:** No path counts, stores or makes an install ID when the endpoint is empty, when `DO_NOT_TRACK=1`, when `ACTA_TELEMETRY` is `0|false|off|no`, or when the setting is off; list each path checked. Keys outside the fixed lists for `commands`, `sessions`, `os` and `arch` become `other`, a skill outside the acta list is dropped, `chat_language` maps to the fixed language list or `other`, a missing setting is `unset`. Every count stops at 100000. `Off` deletes unsent days and the install ID but keeps the notice mark; `On` after `Off` makes a new 32-hex ID. Concurrent `Add` from many goroutines and from two processes leaves every day file parseable; a file that fails to parse is dropped, never returned by `Days`. More than 7 unsent days keeps only the newest 7. Files and folders are owner-only.
- [x] Red: `store_test.go` fails because package `internal/telemetry` does not exist.
- [x] Change: add the store with `state.json` (install ID, off flag, notice mark, last flush time) and one locked, atomically written file per local day under `days/`.
- [x] Commit: feat(telemetry): count usage per local day under ~/.acta/telemetry

### Task 2: Flush and receipts

**Files:** `internal/telemetry/flush.go` (new), `internal/telemetry/flush_test.go` (new)
**Seam:** `telemetry.Flush(ctx, store, client)` and `telemetry.Payload(day)` against an `httptest` server
**verify:** Only finished days are sent, never today, one POST each with the exact body of the spec's request contract and no other field. A day is deleted only on `ok: true` echoing the same `install_id` and `day`, or on `ok: false` with `retryable: false`; a wrong ID, wrong day, missing body, malformed JSON, `retryable: true`, a non-200 status, a timeout and a body over 16 KiB all keep the day. A redirect to anything but the expected HTTPS Google content host as a GET is refused. A second flush that finds the flush lock held returns at once. Every attempt writes the last flush time. Try adversarial receipts: extra fields, `ok` as a string, a huge body, a redirect to plain http.
- [x] Red: `flush_test.go` fails because `Flush` and `Payload` are not defined.
- [x] Change: add the payload builder, the client with a 20-second timeout and the redirect rule, and the receipt handling.
- [x] Commit: feat(telemetry): send finished days and honour collector receipts

### Task 3: acta telemetry command

**Files:** `internal/cli/telemetry.go` (new), `internal/cli/telemetry_test.go` (new), `internal/cli/cli.go`, `internal/cli/help.go`
**Seam:** `cli.Run([]string{"telemetry", ...})`
**verify:** Wiring only, over the store and flush cases above. `on` and `off` change the stored setting; `status` prints the state, its source (setting, environment, or no endpoint in this build), the exact request body of each unsent day and today's counts marked not finished; `flush` runs one flush and prints nothing. A bad subcommand prints the usage line and exits with bad input. Help lists the command.
- [x] Red: `telemetry_test.go` fails because `acta telemetry` is an unknown command.
- [x] Change: add the `telemetry` case to `Run`, the subcommands and the help entry.
- [x] Commit: feat(cli): add acta telemetry on, off, status and flush

### Task 4: Count every command, notice and flush trigger

**Files:** `internal/cli/cli.go`, `internal/cli/telemetry_run.go` (new), `internal/cli/telemetry_run_test.go` (new)
**Seam:** `cli.Run` with a fake stderr TTY flag and a fake flush starter
**verify:** Every call of `Run` counts its top-level command name once, `tui` for no arguments, never an argument or flag, and records the settings snapshot from `config.ResolveUser`. The stderr notice prints only for a command that is not `hook` and whose stderr is a terminal, at most once per machine, and only that path or session-start sets the mark. A command that is not `hook` starts the detached flush only when the last attempt is over one hour old; `hook` commands never start it here. Help and version stay as fast as now and are counted the same way. List every command path checked, the TUI path included.
- [x] Red: `telemetry_run_test.go` fails because `Run` counts nothing.
- [x] Change: count at the top of `Run`, print the notice, and start `acta telemetry flush` as a detached process with output thrown away behind the hourly gate.
- [x] Commit: feat(cli): count commands, show the telemetry notice once, start the hourly flush

## Fix round 1

### Task 5: Fix round 1: real terminal check, crash-proof day files, honest counts

**verify:** No path marks the notice shown unless stderr is a real terminal (not /dev/null, a pipe or a file), no day file shape can crash a command, the background flush never counts itself, a no-endpoint build writes no file under ~/.acta/telemetry, a flush stops sending once telemetry is turned off, and an HTTPS redirect to a host other than script.googleusercontent.com or a redirect that keeps POST is refused; list each path checked.

- [x] B1: isTerminal (internal/cli/notice.go:136) treats /dev/null as a terminal, so `acta --version 2>/dev/null` marks the telemetry notice shown while nobody sees it; use a real tty check, which also fixes the update notice.
- [x] [fix] A day file that parses with missing or null maps makes Add and SetSettings panic on every command (internal/telemetry/store.go readDay); allocate missing maps or drop the file.
- [x] [fix] The hourly `acta telemetry flush` child counts itself as commands.telemetry (internal/cli/telemetry_run.go countRun); do not count `telemetry flush`.
- [x] [fix] A no-endpoint build creates ~/.acta/telemetry/lock and state.json on telemetry on, off and status (internal/telemetry/store.go withLock); write nothing when Endpoint is empty.
- [x] [fix] Flush checks Enabled once, so days loaded before `acta telemetry off` are still sent (internal/telemetry/flush.go); check again before each send.
- [x] [fix] The redirect test covers only plain http (internal/telemetry/flush_test.go); add an HTTPS redirect to another host and a 307 that keeps POST.
- [x] [fix] The no-argument TUI path is not in the counted command paths (internal/cli/telemetry_run_test.go); test commandName for no args and for flags only.
- [x] Commit with the trailer `Task: PLN-bymvb60#5`

## Fix round 2

### Task 6: Fix round 2: off and status still work on a dev build

**verify:** With no endpoint, no command creates ~/.acta/telemetry, yet when the folder already exists acta telemetry off still deletes every day file and the install ID and sets off, and status lists the stored days; and a redirect that keeps POST to script.googleusercontent.com is refused; list each path checked.

- [x] R2-B1: withLock (internal/telemetry/store.go) returns nil with no endpoint, so on a dev build `acta telemetry off` leaves days and the install ID a release build stored, and status says Nothing stored; skip only when the folder does not exist yet.
- [x] [fix] The 307 subtest redirects to another host, so it cannot catch a removed POST check (internal/telemetry/flush_test.go); send the 307 to script.googleusercontent.com through a transport that dials the test server.
- [x] Commit with the trailer `Task: PLN-bymvb60#6`

## Review notes

- Round 1 (deep): BLOCKER 1 (isTerminal treats /dev/null as a terminal) -> fixed in fix round 1. Other findings: strict receipt rules (duplicate required, unknown fields rejected) NOTE open, intended by the plan; usage and telemetry skip the flush NOTE open, intended; no https check on Endpoint NOTE open; broken state.json then on shows the notice again NOTE open; chat_language has no alias map NOTE open; withLock can wait up to 2s NOTE open; hook path cost not measured NOTE open; update notice /dev/null flaw fixed by the same isTerminal change; TestTickWarningChangesNothingElse flake sent to bug.
- Round 2 (deep): R2-B1 (withLock skipped off and status with no endpoint) -> fixed in fix round 2; 307 test could not catch a removed POST check -> fixed in fix round 2. NOTE open: isTerminal has no true-case test; off then on during one flush can still send old-ID days; isTerminal Fd() may switch a non-blocking stderr to blocking; readDay does not fill Kind, Version, OS, Arch on a hand-edited file; a hand-edited day can carry keys outside the fixed lists (pre-existing in this branch, not filtered on send).
- Level deep (review_depth), deep lens for the HTTP send, consent and delete paths. BLOCKER 2, NOTE 29, pre-existing 4, over rounds 1 to 3. Round 3 CLEAN on both axes. No [fix] or [debt] left: open notes are a silent no-op for off on a dev build with no folder yet (plan choice), any Stat error read as no folder, no CLI-level test for off on a dev build over an existing folder, an unguarded slice in one redirect test, plus the round 1 and 2 NOTE-open items.

## State

### Next

Task 2: flush.go uses Store.Days/Delete/SetLastFlush and Endpoint; payload is the Day struct as JSON (Kind already acta_telemetry).
