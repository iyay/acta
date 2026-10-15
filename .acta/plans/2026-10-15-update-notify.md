---
parent: specs/2026-10-15-update-notify-design
depth: minimal
id: PLN-0142
created: "2026-10-15 09:34:51"
hash: u1b98a8
started: "2026-10-15 09:36:30"
finished: "2026-10-15 09:41:22"
---
# Update notice Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Every acta run learns about a newer release through a daily background check and tells the user on CLI stderr, the TUI footer and the SessionStart note, without installing anything.

**Spec:** .acta/specs/2026-10-15-update-notify-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- No run ever waits on the network: the check runs only in the detached `acta update --check --cache` child.
- Stale means cache file mtime older than 24 hours. Opt-outs: source build, `ACTA_NO_UPDATE_CHECK=1`, `update_check: off`.
- Notice copy, verbatim: CLI `acta v<latest> available (you have v<current>). Run: acta update`; TUI `↑ v<latest>`; hook `acta update available: v<latest> (you have v<current>). Tell the user once, in a single line, to run acta update.`
- Tests that touch the cache use a temp user dir, never the real `~/.acta/`.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2, Task 3
- Wave 3: Task 4

### Task 1: Cache, stale check and version compare

**Files:**
- Create: `internal/update/notice.go`, `internal/update/notice_test.go`
- Modify: `internal/config/user.go` (field `UpdateCheck string yaml:"update_check,omitempty"`, valid values empty, `on`, `off`), `internal/config/user_test.go`

**verify:** `Newer(cached, current)` is true only when both parse as `vX.Y.Z` and cached is strictly higher; list every input shape tried (equal, lower, higher in each part, missing `v`, garbage, empty). `ReadCache` on a missing, empty or corrupt file returns no version and no panic. `WriteCache` never leaves a half-written file (temp plus rename). `Stale` is true for a missing file and for mtime older than 24 hours, false otherwise.

- [x] Write failing tests, watch them fail, implement, watch them pass, commit.

### Task 2: CLI: spawn, --cache, stderr notice, update clears it

**Files:**
- Modify: `internal/cli/cli.go`, `internal/cli/update_cmd.go`
- Create: `internal/cli/notice.go`, `internal/cli/notice_test.go`

**verify:** No path through `cli.Run` starts more than one check per stale cache, and none starts one for a source build, `ACTA_NO_UPDATE_CHECK=1`, `update_check: off` or the `update` command; list every path checked. The cache file is touched before the spawn. The stderr line never prints for `acta hook ...`, `--json`, non-terminal stderr, or when no newer version is cached. `update --check --cache` writes the cache on success and prints nothing and leaves the cache as it was on any failure (fake server). A successful `acta update` writes the new version to the cache. The spawn and the terminal check go through variables a test swaps.

- [x] Write failing tests, watch them fail, implement, watch them pass, commit.

### Task 3: SessionStart line

**Files:**
- Modify: `internal/hook/hook.go` (new `Input` field `Update string`, the newer version or empty), `internal/hook/hook_test.go`, `internal/cli/hook.go`

**verify:** The SessionStart text holds the hook line exactly once when `Update` is set and never when it is empty; the prompt hook never holds it. `internal/cli/hook.go` is wiring only: one test with a newer cached version, one with none.

- [x] Write failing tests, watch them fail, implement, watch them pass, commit.

### Task 4: TUI footer badge

**Files:**
- Modify: `internal/tui/title.go` (`WithUpdate(v string) Model`), its test file, `internal/cli/cli.go` (pass the cached newer version when building the model)

**verify:** The footer shows `↑ v<latest>` only when `WithUpdate` got a non-empty version, and the footer never grows wider than the screen at any width the existing footer tests sweep; list the widths checked.

- [x] Write failing tests, watch them fail, implement, watch them pass, commit.

## Review notes

- Round 1, deep, two reviewers (spec, standards): both CLEAN, no BLOCKER.
- Hooks also start the daily check; the spec skip list names only `update`, and the per-prompt cost is one config read and one stat.
- A failed background check leaves an empty touched `update.json`; the next try comes 24 hours later, as the spec says.
- The badge test lives in `view_test.go` beside the other footer tests, following the approved footer placement.
