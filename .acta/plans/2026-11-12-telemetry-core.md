---
parent: specs/2026-11-12-usage-telemetry-design
depth: minimal
id: PLN-0228
created: "2026-11-12 06:31:58"
hash: bymvb60
started: "2026-11-12 06:35:55"
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
- [ ] Red: `telemetry_test.go` fails because `acta telemetry` is an unknown command.
- [ ] Change: add the `telemetry` case to `Run`, the subcommands and the help entry.
- [ ] Commit: feat(cli): add acta telemetry on, off, status and flush

### Task 4: Count every command, notice and flush trigger

**Files:** `internal/cli/cli.go`, `internal/cli/telemetry_run.go` (new), `internal/cli/telemetry_run_test.go` (new)
**Seam:** `cli.Run` with a fake stderr TTY flag and a fake flush starter
**verify:** Every call of `Run` counts its top-level command name once, `tui` for no arguments, never an argument or flag, and records the settings snapshot from `config.ResolveUser`. The stderr notice prints only for a command that is not `hook` and whose stderr is a terminal, at most once per machine, and only that path or session-start sets the mark. A command that is not `hook` starts the detached flush only when the last attempt is over one hour old; `hook` commands never start it here. Help and version stay as fast as now and are counted the same way. List every command path checked, the TUI path included.
- [ ] Red: `telemetry_run_test.go` fails because `Run` counts nothing.
- [ ] Change: count at the top of `Run`, print the notice, and start `acta telemetry flush` as a detached process with output thrown away behind the hourly gate.
- [ ] Commit: feat(cli): count commands, show the telemetry notice once, start the hourly flush

## State

### Next

Task 2: flush.go uses Store.Days/Delete/SetLastFlush and Endpoint; payload is the Day struct as JSON (Kind already acta_telemetry).
