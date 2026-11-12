---
parent: specs/2026-11-12-usage-telemetry-design
depth: minimal
id: PLN-0229
created: "2026-11-12 06:31:58"
hash: lr1abyy
after: [PLN-0228]
started: "2026-11-12 07:30:50"
finished: "2026-11-12 07:40:49"
---
# Telemetry from hooks, omp and workflow commands

**Goal:** Sessions per harness, acta skill loads, pre-tool blocks and spec, plan, bug and land events reach the telemetry store, and the session-start hook shows the notice line and starts the hourly flush.

**Spec:** `.acta/specs/2026-11-12-usage-telemetry-design.md` (SPC-0201)

**Tests:** fast `scripts/test ./internal/cli ./internal/hook` and `cd plugin/omp && bun test`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Uses the `internal/telemetry` store from the core plan; adds no rule of its own about what is stored.
- A hook never waits on telemetry, never prints telemetry text to stderr, and keeps its exit code and block message word for word.
- Tests set `HOME` to `t.TempDir()` and swap `telemetry.Endpoint`; none reaches the network.

## Waves

- Wave 1: Task 1, Task 3, Task 4
- Wave 2: Task 2

### Task 1: Session-start counts the harness, shows the notice, starts the flush

**Files:** `internal/cli/hook.go`, `internal/cli/hook_telemetry_test.go` (new), `internal/hook/hook.go`, `internal/hook/hook_test.go`
**Seam:** `cli.Run([]string{"hook", "session-start"})` with env and stdin set per harness
**verify:** Each session-start counts one session: `omp` when `ACTA_HARNESS=omp`, `codex` when `hook.IsCodex`, `claude` otherwise, `other` for any other `ACTA_HARNESS` value. The notice line appears in the hook text only while the notice is not yet marked, and printing it marks it; with telemetry off or no endpoint the line never appears. The hook starts the detached flush only past the hourly gate and returns without waiting. The 60-line cap test still passes.
- [x] Red: `hook_telemetry_test.go` fails because session-start counts no session.
- [x] Change: count the harness, add the one notice line to the session text, and start the gated flush.
- [x] Commit: feat(hook): count sessions per harness and show the telemetry notice at session start

### Task 2: Pre-tool counts acta skills and blocks

**Files:** `internal/cli/hook.go`, `internal/cli/hook_telemetry_test.go`, `plugin/hooks/hooks.json`, `internal/plugincheck/plugin_test.go`
**Seam:** `cli.Run([]string{"hook", "pre-tool"})` with Claude Code and omp shaped stdin
**verify:** A Claude Code `Skill` call with `tool_input.skill` `acta:<name>` and an omp read of `skill://<name>` each count `<name>` once when it is in the acta skill list; any other skill name, with or without a prefix, records nothing. Each `go_test` and `destructive` block counts its rule once and keeps the exact block text and exit 2. A `Skill` call makes no wiki hint and no block. `hooks.json` lists `Skill` in the PreToolUse matcher. Try adversarial names: `acta:../x`, `acta:`, a 10 KB name, mixed case.
- [x] Red: `hook_telemetry_test.go` fails because a `Skill` call and a block count nothing.
- [x] Change: read the skill name in pre-tool, count blocks where they are decided, and add `Skill` to the matcher.
- [x] Commit: feat(hook): count acta skill loads and pre-tool blocks

### Task 3: omp extension names its harness and forwards skill reads

**Files:** `plugin/omp/index.ts`, `plugin/omp/index.test.ts`
**Seam:** the extension's `realRun` and its tool-event handler, with `spawnSync` faked
**verify:** Every acta call the extension makes carries `ACTA_HARNESS=omp` in its env and keeps the rest of the env. A read of `skill://<name>` is forwarded to `acta hook pre-tool` with the same payload shape as a file read; other reads behave as before.
- [x] Red: `index.test.ts` fails because the spawned env has no `ACTA_HARNESS`.
- [x] Change: add `ACTA_HARNESS: "omp"` to the spawn env and forward `skill://` reads.
- [x] Commit: feat(omp): tell acta the harness and forward skill reads

### Task 4: Workflow counters

**Files:** `internal/cli/id.go`, `internal/cli/specdates.go`, `internal/cli/cli.go`, `internal/cli/workflow_telemetry_test.go` (new)
**Seam:** `cli.Run` for `id`, `bug new` and `spec-dates`, in a temp clone
**verify:** `spec_created` and `plan_created` go up once for each new SPC or PLN number `acta id` gives, and never for an item that already had one or for `--fix-duplicates` renumbering. `bug_created` goes up once per bug `acta bug new` writes, and not when it fails. `plan_landed` goes up once per `acta spec-dates <plan hash>` run that succeeds. List every path checked.
- [x] Red: `workflow_telemetry_test.go` fails because these commands count nothing.
- [x] Change: count after each successful write in the three commands.
- [x] Commit: feat(cli): count specs, plans, bugs and landed plans

## Review notes

- Level deep (review_depth), deep lens on hook stdin. BLOCKER 0, NOTE 13, pre-existing 3 (ParseEvent drops payloads with no session_id; Flock and Setsid do not build on Windows, SCR-0058; one locked disk write per pre-tool call), round 1 CLEAN on both axes. The Skill early-return test gap stays a note: it sits on the pre-tool block path, so it is not a polish fix, and no real Skill payload carries a command. Open notes: sessions count clear, compact and resume starts too; the notice is marked before it prints; a /acta:build slash run may skip the Skill tool; the omp child_process mock is never undone; a non-acta omp skill named like an acta one counts as acta; Codex needs codex plugin add again for the new matcher.

## State

### Next

SPC-0201 parallel run
PLN-0228 landed
PLN-0229 review CLEAN, waiting to land after PLN-0230
PLN-0230 review CLEAN, land gate running
